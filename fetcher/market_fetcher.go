package fetcher

import (
	"cofin/integrations"
	"cofin/internal"
	"cofin/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MarketFetcher struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewMarketFetcher(db *gorm.DB) (*MarketFetcher, error) {
	logger, err := internal.NewLogger()
	if err != nil {
		return nil, err
	}

	return &MarketFetcher{
		db:     db,
		logger: logger,
	}, nil
}

func (f *MarketFetcher) Run() {
	logger := f.logger
	db := f.db

	fetchMarket(db, logger)
}

func fetchMarket(db *gorm.DB, logger *zap.SugaredLogger) {
	logger.Info("Running fetching job...")

	var companies []models.Company
	result := db.Find(&companies)
	if result.Error != nil {
		logger.Errorf("Failed to fetch list of companies from database: %v", result.Error)
		return
	}

	for _, company := range companies {
		logger.Infof("Fetching data for %v", company.Ticker)

		realStonks := integrations.RealStonks{}
		marketInformation, err := realStonks.GetMarketData(company.Ticker)
		if err == nil {
			company.Currency = marketInformation.Currency
			company.Price = marketInformation.Price
			company.Change = marketInformation.Change
			company.TotalVolume = marketInformation.TotalVolume

			tx := db.Save(company)
			if tx.Error != nil {
				logger.Infof("Unable to update market data for %v: %v", company.Ticker, tx.Error)
			}
		} else {
			logger.Infof("Unable to fetch market data for %v", company.Ticker)
		}
	}

}
