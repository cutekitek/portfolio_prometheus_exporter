package tinkoff

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	"github.com/cutekitek/portfolio_prometheus_exporter/config"
	"github.com/cutekitek/portfolio_prometheus_exporter/metrics"
	pb "github.com/tinkoff/invest-api-go-sdk/proto"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
)

type TinkoffClient struct {
	conn     *grpc.ClientConn
	metrics  metrics.MetricsManager
	logger   *slog.Logger
	apiKey   string
	appName  string
	interval time.Duration

	usersService      pb.UsersServiceClient
	operationsService pb.OperationsServiceClient
	instrumentsService pb.InstrumentsServiceClient
}

type authKey string

const (
	defaultApiEndpoint = "invest-public-api.tinkoff.ru:443"
)

func NewTinkoffExporter(ctx context.Context, cfg config.ExchangeConfig, metricsManager metrics.MetricsManager) (*TinkoffClient, error) {
	apiKey, ok := cfg.Params["api_key"]
	if !ok {
		return nil, fmt.Errorf("failed to initialize tinkoff exporter: missing api_key")
	}
	appName, ok := cfg.Params["app_name"]
	if !ok {
		return nil, fmt.Errorf("failed to initialize tinkoff exporter: missing app_name")
	}
	endpoint, ok := cfg.Params["api_endpoint"]
	if !ok {
		endpoint = defaultApiEndpoint
	}

	conn, err := grpc.Dial(endpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(oauth.TokenSource{
			TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey}),
		}))

	if err != nil {
		return nil, fmt.Errorf("failed to initialize tinkoff exporter:%w", err)
	}

	return &TinkoffClient{
		conn:              conn,
		logger:            slog.With("exchange", "tinkoff"),
		apiKey:            apiKey,
		appName:           appName,
		metrics:           metricsManager,
		interval:          cfg.ScrapingInterval,
		usersService:      pb.NewUsersServiceClient(conn),
		operationsService: pb.NewOperationsServiceClient(conn),
		instrumentsService: pb.NewInstrumentsServiceClient(conn),
	}, nil
}

func (c *TinkoffClient) Scraper(ctx context.Context) {
	ctx = context.WithValue(ctx, authKey("authorization"), fmt.Sprintf("Bearer %s", c.apiKey))
	ctx = metadata.AppendToOutgoingContext(ctx, "x-app-name", c.appName)
	c.logger.Info("starting exporter", "interval", c.interval.Seconds())
	t := time.NewTicker(c.interval)
	for {
		select {
		case <-t.C:
			c.process(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (c *TinkoffClient) process(ctx context.Context) {
	portfolios, err := c.usersService.GetAccounts(ctx, &pb.GetAccountsRequest{})
	if err != nil {
		c.logger.Error("failed to get portfolios", "error", err)
		return
	}

	acc := portfolios.GetAccounts()
	if len(acc) == 0{
		c.logger.Error("portfolios empty", "error", err)
		return
	}
	p := acc[0]


	balance, err := c.operationsService.GetPortfolio(ctx, &pb.PortfolioRequest{AccountId: p.Id, Currency: pb.PortfolioRequest_RUB})
	if err != nil {
		c.logger.Error("failed to get portfolio balance", "error", err, "portfolio", p.Name)
		return
	}
	
	c.metrics.UpdateAccountMoneyValue("RUB", balance.TotalAmountPortfolio.ToFloat())
	c.metrics.UpdateAccountRelativeYield(balance.ExpectedYield.ToFloat())
	for _, share := range balance.GetPositions() {
		shareData, err := c.instrumentsService.GetInstrumentBy(ctx, &pb.InstrumentRequest{IdType: pb.InstrumentIdType_INSTRUMENT_ID_TYPE_UID, Id: share.InstrumentUid})
		if err != nil{
			c.logger.Error("failed to get share data", "error", err)
			continue
		}
		
		c.metrics.UpdateShareRelativeYield(shareData.Instrument.Ticker, share.ExpectedYield.ToFloat())
		c.metrics.UpdateAccountShareCount(shareData.Instrument.Ticker, share.Quantity.ToFloat())
		c.metrics.UpdateSharePrice(shareData.Instrument.Ticker, share.CurrentPrice.ToFloat() * float64(shareData.Instrument.Lot))
	}
}
