package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	_ "embed"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
)

const (
	dateFormat        = "Jan 2, 2006"
	receiptDateFormat = "January 2006"
	stdDateFormat     = "02-01-2006"
)

var (
	//go:embed receipt.html
	receiptTemplate string
)

type Receipt struct {
	ID            int
	Tenant        Person
	Landlord      Person
	Amount        uint64
	startDate     time.Time
	endDate       time.Time
	issueDate     time.Time
	IsProvisional bool
}

func (r *Receipt) StartDate() string {
	return r.startDate.Format(dateFormat)
}
func (r *Receipt) IssueDate() string {
	return r.issueDate.Format(dateFormat)
}

func (r *Receipt) EndDate() string {
	return r.endDate.Format(dateFormat)
}

func (r *Receipt) ReceiptDate() string {
	return r.startDate.Format(receiptDateFormat)
}

func main() {
	var cfgFile string
	flag.StringVar(&cfgFile, "config", "", "specify the config file to be used")
	flag.Parse()
	if cfgFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	config, err := ParseConfig(cfgFile)
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	if err != nil {
		zap.L().Fatal("failed parse config yaml", zap.Error(err))
	}

	zap.L().Info("loaded config file", zap.Any("config", config))

	fyStart := time.Date(int(config.FinancialYear), 4, 1, 0, 0, 0, 0, time.Local)
	var receipts []*Receipt
	for i := 0; i < 12; i++ {
		receipt := &Receipt{
			Tenant:    config.Tenant,
			Landlord:  config.Landlord,
			Amount:    config.Rent,
			ID:        i + 1,
			startDate: fyStart.AddDate(0, i, 0),
			endDate:   fyStart.AddDate(0, i+1, -1),
			issueDate: fyStart.AddDate(0, i+1, 0),
		}
		if receipt.issueDate.Month() == time.April {
			receipt.issueDate = receipt.issueDate.AddDate(0, -1, 0)
		}
		receipts = append(receipts, receipt)
	}
	generatePDF(receipts, config)
}

func writeHTML(receipts []*Receipt) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		tpl, err := template.New("receipt").Parse(receiptTemplate)
		if err != nil {
			zap.L().Fatal("failed to parse template", zap.Error(err))
		}

		if err := tpl.Execute(w, receipts); err != nil {
			zap.L().Fatal("failed to generate receipts", zap.Error(err))
		}
		w.WriteHeader(http.StatusOK)
	})
}

// generatePDF creates a PDF with the given data
func generatePDF(receipts []*Receipt, config *Config) (string, error) {
	file := fmt.Sprintf("./rent_receipt_%d.pdf", config.FinancialYear)
	// Convert objects and save the output PDF document.
	outFile, err := os.Create(file)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ts := httptest.NewServer(writeHTML(receipts))

	defer ts.Close()

	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, _, err = page.PrintToPDF().
				WithDisplayHeaderFooter(false).
				WithPaperHeight(11.7).
				WithPaperWidth(8.27).
				WithLandscape(false).
				Do(ctx)
			return err
		}),
	); err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(file, buf, 0644); err != nil {
		return "", err
	}
	return file, nil
}
