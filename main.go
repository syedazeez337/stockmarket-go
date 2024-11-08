package main

import (
    "context"
    "encoding/csv"
    "fmt"
    "log"
    "math/rand"
    "os"
    "time"

    "github.com/chromedp/chromedp"
)

func main() {
    ticker := []string{"MSFT", "IBM", "AAPL"}
    stocks := []Stock{}

    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.UserAgent(randomUserAgent()),
        chromedp.Flag("headless", true),
        chromedp.Flag("disable-gpu", true),
    )

    ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    defer cancel()

    ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(func(s string, args ...interface{}) {
        // Filter out specific unhandled page events
        if s == "ERROR: unhandled page event *page.EventFrameSubtreeWillBeDetached" {
            return
        }
        log.Printf(s, args...)
    }))
    defer cancel()

    for _, t := range ticker {
        var company, price, change string
        url := "https://finance.yahoo.com/quote/" + t

        err := chromedp.Run(ctx,
            chromedp.Navigate(url),
            chromedp.Sleep(randomDelay()),
            chromedp.Text(`h1`, &company, chromedp.NodeVisible),
            chromedp.Text(`[data-field="regularMarketPrice"]`, &price, chromedp.NodeVisible),
            chromedp.Text(`[data-field="regularMarketChangePercent"]`, &change, chromedp.NodeVisible),
        )
        if err != nil {
            log.Println("Error fetching data for", t, ":", err)
            continue
        }

        stocks = append(stocks, Stock{Company: company, Price: price, Change: change})
        fmt.Printf("Fetched data for %s: %s, %s, %s\n", t, company, price, change)
    }

    writeCSV(stocks)
}

type Stock struct {
    Company string
    Price   string
    Change  string
}

func writeCSV(stocks []Stock) {
    file, err := os.Create("stocks.csv")
    if err != nil {
        log.Fatalln("Failed to create output CSV file", err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    headers := []string{"Company", "Price", "Change"}
    writer.Write(headers)

    for _, stock := range stocks {
        record := []string{stock.Company, stock.Price, stock.Change}
        writer.Write(record)
    }

    writer.Flush()
    if err := writer.Error(); err != nil {
        log.Fatalf("Error flushing writer: %v", err)
    }
}

func randomUserAgent() string {
    userAgents := []string{
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Safari/605.1.15",
        "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.181 Mobile Safari/537.36",
    }
    rand.Seed(time.Now().UnixNano())
    return userAgents[rand.Intn(len(userAgents))]
}

func randomDelay() time.Duration {
    return time.Duration(rand.Intn(3)+1) * time.Second
}
