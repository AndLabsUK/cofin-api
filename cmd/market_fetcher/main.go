package main

import (
	"cofin/core"
	"cofin/internal/real_stonks"
	"cofin/models"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()

	// connect to the database
	db, err := core.InitDB()
	if err != nil {
		panic(err)
	}

	// auto migrate the database
	err = db.Debug().AutoMigrate(
		&models.User{},
		&models.Company{},
		&models.Document{},
		&models.AccessToken{},
	)
	if err != nil {
		panic(err)
	}

	fetcher, err := NewMarketFetcher(db)
	if err != nil {
		panic(err)
	}

	fetcher.Run()
}

type MarketFetcher struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewMarketFetcher(db *gorm.DB) (*MarketFetcher, error) {
	logger, err := core.NewLogger()
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

		realStonks := real_stonks.RealStonks{}
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
