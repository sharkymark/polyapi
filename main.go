package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
    "bufio"
    "os"
    "strconv"
    "log"
    "time"
    "io"
    "database/sql"
	_ "github.com/mattn/go-sqlite3"
	
)

type GeoCodingResponse struct {
	Result struct {
		AddressMatches []struct {
			Coordinates struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"coordinates"`
            MatchedAddress string `json:"matchedAddress"`
		} `json:"addressMatches"`
	} `json:"result"`
}

type NOAAWeatherResponse struct {
    Properties struct {
        Forecast          string `json:"forecast"`
        ForecastHourly    string `json:"forecastHourly"`
        ForecastGridData  string `json:"forecastGridData"`
        ObservationStations string `json:"observationStations"`
        ForecastZone      string `json:"forecastZone"`
        County            string `json:"county"`
        FireWeatherZone   string `json:"fireWeatherZone"`
        Periods []struct {
            Number          int    `json:"number"`
            Name            string `json:"name"`
            StartTime       string `json:"startTime"`
            EndTime         string `json:"endTime"`
            IsDaytime       bool   `json:"isDaytime"`
            Temperature     int    `json:"temperature"`
            TemperatureUnit string `json:"temperatureUnit"`
            ShortForecast   string `json:"shortForecast"`
            DetailedForecast string `json:"detailedForecast"`
        } `json:"periods"`
    } `json:"properties"`
}

type TreasuryData struct {
        RecordDate         string     `json:"record_date"`
        SecurityTypeDesc   string     `json:"security_type_desc"`
        SecurityDesc       string     `json:"security_desc"`
        AvgInterestRateAmt string     `json:"avg_interest_rate_amt"`
        SrcLineNbr         string      `json:"src_line_nbr"`
}

type TreasuryResponse struct {
    Data []TreasuryData `json:"data"`
} 


type Address struct {
    Id              int
    MatchedAddress  string
    Latitude        float64
    Longitude       float64
    LastTemperature string
    UpdatedAt       string
}

type Ticker struct {
    CompanyName string
    Ticker      string
    Exchange    string
    LastPrice   string
    UpdatedAt   string
    Id          int
}

// deleteAddress deletes an address from the database.
func deleteAddress(db *sql.DB, id int) {

    // Delete the address from the database
    result, err := db.Exec("DELETE FROM addresses WHERE id = ?", id)
    if err != nil {
        log.Fatal(err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println()

    if rowsAffected > 0 {
        fmt.Println("Address deleted successfully.")
    } else {
        fmt.Println("Address not found.")
    }
}

// deleteTicker deletes a ticker symbol from the database.
func deleteTicker(db *sql.DB, id int) {

    // Delete the ticker from the database
    result, err := db.Exec("DELETE FROM tickers WHERE id = ?", id)
    if err != nil {
        log.Fatal(err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println()

    if rowsAffected > 0 {
        fmt.Println("Ticker symbol deleted successfully.")
    } else {
        fmt.Println("Ticker symbol not found.")
    }
}

/*
func addColumn(db *sql.DB, tableName, columnName, columnType string) error {
    query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
    println(query)
    
    _, err := db.Exec(query)
    if err != nil {
        return fmt.Errorf("error adding column to table: %w", err)
    }
    return nil
}
*/

func addColumn(db *sql.DB, tableName, columnName, columnType string) error {
    // Check if column exists
    query := fmt.Sprintf("SELECT name FROM pragma_table_info('%s') WHERE name = '%s'", tableName, columnName)
    row := db.QueryRow(query)
    var existingName string
    err := row.Scan(&existingName)
    if err != nil && err != sql.ErrNoRows { // Ignore "no rows" error
        return fmt.Errorf("error checking for existing column: %w", err)
    }

    if existingName == "" { // Column doesn't exist
        query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType)
        _, err := db.Exec(query)
        if err != nil {
            return fmt.Errorf("error adding column to table: %w", err)
        }
    }
    return nil
}


func createDB() *sql.DB{
	dbFile := "./db/polyapi.db"
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		// Create the database file
		_, err := os.Create(dbFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	// Create tables for storing address and ticker symbol data
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS addresses (
			id INTEGER PRIMARY KEY,
			address TEXT NOT NULL,
			lat REAL NOT NULL,
			lon REAL NOT NULL,
            last_temperature TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS tickers (
			id INTEGER PRIMARY KEY,
			ticker TEXT NOT NULL,
			company_name TEXT NOT NULL,
            sector TEXT NOT NULL,
            industry TEXT NOT NULL,
            exchange TEXT NOT NULL,
            address TEXT NOT NULL,
            official_site TEXT NOT NULL,
            revenue_ttm REAL NOT NULL,
            market_cap REAL NOT NULL,
            fiscal_year_end TEXT NOT NULL,
            last_price REAL NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

    // Check if tables require additional columns
    // Add updated_at column to addresses table
    err = addColumn(db, "addresses", "updated_at", "TIMESTAMP")
    if err != nil {
        log.Fatal("Error adding updated_at to addresses table:", err)
    }

    // Add updated_at column to tickers table
    err = addColumn(db, "tickers", "updated_at", "TIMESTAMP")
    if err != nil {
        log.Fatal("Error adding updated_at to tickers table:", err)
    }

    return db

}

// printForecast prints the weather forecast for a location.
// Specifically, it prints the next 4 periods and the last 2 periods.
// A period is a 12-hour time frame.
func printForecast(noaaResponse NOAAWeatherResponse) {

    // Call the forecast API
    resp, err := http.Get(noaaResponse.Properties.Forecast)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return
    }

    //fmt.Println(string(body))

    err = json.Unmarshal(body, &noaaResponse)
    if err != nil {
        fmt.Println(err)
        return
    }

    println("\nForecast: (next 2 days and a week out)")
    println()

    // Check if Periods array is not empty
    if len(noaaResponse.Properties.Periods) == 0 {
        fmt.Println("No forecast periods available.")
        return
    }

    // Print the next 4 periods
    for i := 0; i < 4; i++ {
        period := noaaResponse.Properties.Periods[i]
        startTime := strings.Split(period.StartTime, "T")[0]
        fmt.Printf("%s (%s)  %d%s\n", startTime, period.Name, period.Temperature, period.TemperatureUnit)
        fmt.Printf("  %s\n\n", period.DetailedForecast)
    }

    // Print the last 2 periods
    for i := len(noaaResponse.Properties.Periods) - 2; i < len(noaaResponse.Properties.Periods); i++ {
        period := noaaResponse.Properties.Periods[i]
        startTime := strings.Split(period.StartTime, "T")[0]
        fmt.Printf("%s (%s)  %d%s\n", startTime, period.Name, period.Temperature, period.TemperatureUnit)
        fmt.Printf("  %s\n\n", period.DetailedForecast)
    }
}

//  extractDate extracts the date from a timestamp.
func extractDate(timestamp string) string {
    t, _ := time.Parse("2006-01-02T15:04:05", timestamp)
    return t.Format("2006-01-02")
}

// formatTime formats the start time of a period.
func formatTime(startTime string) string {
    startTime = strings.Split(startTime, "T")[1]
    hourMinute := strings.Split(startTime, "-")[0]
    hourMinuteParts := strings.Split(hourMinute, ":")
    hour := hourMinuteParts[0]
    minute := hourMinuteParts[1]
    ampm := "AM"
    hourInt, _ := strconv.Atoi(hour)
    if hourInt >= 12 {
        ampm = "PM"
    }
    if hourInt == 0 {
        hour = "12"
    } else if hourInt > 12 {
        hour = strconv.Itoa(hourInt - 12)
    }
    return fmt.Sprintf("%s:%s %s", hour, minute, ampm)
}


// printHourlyForecast prints the hourly weather forecast for a location.
// Specifically, it prints the next 12 hours.
func printHourlyForecast(noaaResponse NOAAWeatherResponse) {

    // Call the hourly forecast API
    resp, err := http.Get(noaaResponse.Properties.ForecastHourly)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return
    }

    err = json.Unmarshal(body, &noaaResponse)
    if err != nil {
        fmt.Println(err)
        return
    }    

    println("\nNext 12 hours:")
    println()

    for i := 0; i < 12 && i < len(noaaResponse.Properties.Periods); i++ {
        period := noaaResponse.Properties.Periods[i]
        startTime := formatTime(period.StartTime)
        fmt.Printf("%s %d%s\n", startTime, period.Temperature, period.TemperatureUnit)
        fmt.Printf(" - %s\n", period.ShortForecast)
        fmt.Printf("\n")
    }
}

// Get first hourly temperature from the hourly forecast and update the address record in the database.
func updateTemperature(db *sql.DB, noaaResponse NOAAWeatherResponse, addressId int) {

    // Call the hourly forecast API
    resp, err := http.Get(noaaResponse.Properties.ForecastHourly)
    if err != nil {
        fmt.Println(err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
    }

    err = json.Unmarshal(body, &noaaResponse)
    if err != nil {
        fmt.Println(err)
    }

    if len(noaaResponse.Properties.Periods) > 0 {
        firstPeriod := noaaResponse.Properties.Periods[0]
        temperature := strconv.Itoa(firstPeriod.Temperature)
        unit := firstPeriod.TemperatureUnit
        temp_and_unit := temperature + unit
        if temperature != "" {
            // Update address record with temperature
            _, err = db.Exec("UPDATE addresses SET last_temperature = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", temp_and_unit, addressId)
            if err != nil {
                log.Printf("Error updating address record: %v", err)
            } else {
                fmt.Println()
                fmt.Printf("Address record updated successfully with latest temperature %s!\n", temp_and_unit)
            }
        }
      } else {
        // Handle the case where there are no periods available
        fmt.Println("No hourly forecast data available")
      }

}

// getNOAAWeather sends a request to the NOAA API to get the weather forecast for a location.
// it takes the latitude and longitude of the location as arguments.
// it is called from geocode function.

func getNOAAWeather(lat, lon string, db *sql.DB, addressId int) {
    reader := bufio.NewReader(os.Stdin)
    
    // First NOAA API call
    url := fmt.Sprintf("https://api.weather.gov/points/%s,%s", lat, lon)
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return
    }
    var noaaResponse NOAAWeatherResponse

    err = json.Unmarshal(body, &noaaResponse)
    if err != nil {
        fmt.Println(err)
        return
    }
    
    updateTemperature(db, noaaResponse, addressId)

    printHourlyForecast(noaaResponse)

    // Submenu
    for {
        fmt.Println("\nNOAA Weather Submenu:")
        fmt.Println()
        fmt.Println("1. Forecast")
        fmt.Println("2. Hourly Forecast")
        fmt.Println("3. Return to Main Menu")
        fmt.Println()
        
        var option string
        fmt.Print("Enter your option: ")
        option, _ = reader.ReadString('\n')
        option = strings.TrimSpace(option)
        fmt.Println()

        switch option {
        case "1":
            
            printForecast(noaaResponse)

        case "2":

            printHourlyForecast(noaaResponse)

            //fmt.Println(string(body))
        case "3":
            return
        default:
            fmt.Println("\nInvalid option")
        }
    }
}

// getGeoCode sends a request to the Census Geocoding API to get the coordinates of an address.

func getGeoCode(db *sql.DB) {

    reader := bufio.NewReader(os.Stdin)

	// Prompt user for address
	fmt.Print("\nEnter address: (e.g., 432 Park Ave, 10022 or 432 Park Ave NY, NY 10022) [Ctrl+D to quit]\n\n")
    address, err := reader.ReadString('\n')
    if err != nil {
        if err == io.EOF {
            fmt.Println("Cancelled")
            fmt.Println()
            return // Return to previous menu
        }
        fmt.Println("Error reading input:", err)
        return
    }
    address = strings.TrimSpace(address)
    address = strings.ReplaceAll(address, " ", "+")

	// Construct the API request
	url := fmt.Sprintf("https://geocoding.geo.census.gov/geocoder/locations/onelineaddress?address=%s&benchmark=4&format=json", address)

    //println(url)

	// Send the request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}

    //println(resp.Status)
    //println(resp.Body)

	// Read the response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Unmarshal the JSON response
	var response GeoCodingResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the coordinates
	if len(response.Result.AddressMatches) > 0 {
        fmt.Println("\nCoordinates:")
		fmt.Printf("  Latitude: %f\n", response.Result.AddressMatches[0].Coordinates.Y)
		fmt.Printf("  Longitude: %f\n", response.Result.AddressMatches[0].Coordinates.X)

        // Insert new address into database
        matchedAddress := response.Result.AddressMatches[0].MatchedAddress
        var result sql.Result
        result, err = db.Exec("INSERT INTO addresses (address, lat, lon) VALUES (?, ?, ?)", matchedAddress, response.Result.AddressMatches[0].Coordinates.Y, response.Result.AddressMatches[0].Coordinates.X)
        if err != nil {
            log.Fatal(err)
        }

        // Get the ID of the inserted record
        insertedId, err := result.LastInsertId()
        if err != nil {
            log.Printf("Error getting inserted ID: %v", err)
            return // Handle the error appropriately (e.g., log and continue)
        }

        getNOAAWeather(fmt.Sprintf("%f", response.Result.AddressMatches[0].Coordinates.Y), fmt.Sprintf("%f", response.Result.AddressMatches[0].Coordinates.X), db, int(insertedId))
    } else {
        fmt.Println("No coordinates found")
    }
}

//  reuseAddress allows the user to choose a previously entered address from the database.
func reuseAddress(db *sql.DB) {

    // Retrieve unique addresses from the database
    rows, err := db.Query("SELECT id, address, lat, lon, updated_at, last_temperature FROM addresses GROUP BY address")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    var addresses []Address
    for rows.Next() {
        var address Address
        var updatedAt, lastTemperature interface{}
        err := rows.Scan(&address.Id, &address.MatchedAddress, &address.Latitude, &address.Longitude, &updatedAt, &lastTemperature)
        if err != nil {
            log.Fatal(err)
        }
        
        if updatedAt != nil {
            t := updatedAt.(time.Time)
            address.UpdatedAt = t.Format("2006-01-02T15:04:05")
        } else {
            address.UpdatedAt = ""
        }

        if lastTemperature != nil {
            address.LastTemperature = fmt.Sprintf("%v", lastTemperature)
        } else {
            address.LastTemperature = ""
        }

        addresses = append(addresses, address)
    }

    if len(addresses) == 0 {
        fmt.Println("No addresses found")
        return
    }

    fmt.Println("Previous addresses")
    fmt.Println()

    // Assign numbers to each result and ask the user to choose an address
    for i, address := range addresses {
        var updatedat string
        if address.UpdatedAt != "" {
             updatedat = extractDate(address.UpdatedAt)
        }

        var lastTemp string
        if address.LastTemperature != "" {
            lastTemp = address.LastTemperature
        }
        var extraInfo string
        if updatedat != "" || lastTemp != "" {
            extraInfo = fmt.Sprintf("%s on %s", lastTemp, func() string {
                if updatedat != "" {
                    return updatedat
                }
                return ""
            }())
        }
        fmt.Printf("%d. %s ~ %s\n  @ lat: %f, lon: %f\n", i+1, address.MatchedAddress, extraInfo, address.Latitude, address.Longitude)
    }

    fmt.Printf("\nEnter the row number (%d-%d): ", 1, len(addresses))
    var choice int
    _, err = fmt.Scan(&choice)
    if err != nil {
        fmt.Println("Invalid input")
        return
    }

    // Validate user input
    if choice < 1 || choice > len(addresses) {
        fmt.Println("Invalid choice")
        return
    }

    fmt.Println("\n1. Reuse")
    fmt.Println("2. Delete")
    fmt.Println("3. Return to previous menu")
    fmt.Println()
    fmt.Print("Enter your choice: ")
    var action int
    _, err = fmt.Scan(&action)
    if err != nil {
        fmt.Println("Invalid input")
        return
    }

    switch action {
    case 1:
        // Reuse logic using addresses[choice-1]
        chosenAddress := addresses[choice-1]
        lat := chosenAddress.Latitude
        lon := chosenAddress.Longitude
        getNOAAWeather(fmt.Sprintf("%.8f", lat), fmt.Sprintf("%.8f", lon), db, chosenAddress.Id)
    case 2:
        // Delete the selected address
        deleteAddress(db, addresses[choice-1].Id)
    case 3:
        // Return to previous menu
        return
    default:
        fmt.Println("Invalid choice")
    }

}

func geocodeMenu(db *sql.DB) {
    fmt.Println("\nGeocode menu:")
    fmt.Println()
    fmt.Println("1. Enter a new address")
    fmt.Println("2. Re-use/delete a previous address")
    fmt.Println()

    var option int
    fmt.Scanln(&option)

    fmt.Println() 

    switch option {
    case 1:
        // Enter a new address
        getGeoCode(db)
    case 2:
        // Re-use a previous address
        reuseAddress(db)
    }
}

// getStockOverview sends a request to the Alpha Vantage API to get an overview of a company
func getStockOverview(db *sql.DB, tickerSymbol interface{}, lastPrice interface{}, action string) {
    apiKey := os.Getenv("ALPHAVANTAGE_API_KEY")
    if apiKey == "" {
        fmt.Println("ALPHAVANTAGE_API_KEY environment variable is not set.")
        return
    }

    url := fmt.Sprintf("https://www.alphavantage.co/query?function=OVERVIEW&symbol=%s&apikey=%s", tickerSymbol, apiKey)
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer resp.Body.Close()

    /*
    fmt.Println("Response Status:", resp.Status)
    fmt.Println("Response Body:")
    body, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(body))
    */
    var data map[string]string
    err = json.NewDecoder(resp.Body).Decode(&data)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Check for quota exceeded message
    if _, ok := data["Information"]; ok {
        fmt.Println("Daily API quota exceeded. Please refer to Alpha Vantage's premium plans for higher limits.")
        fmt.Println()
        return
    }

    // for first time quote, insert data into database
    // for update, update data in database with last price
    if action == "insert" {

        // Insert data into database
        _, err = db.Exec(`
            INSERT INTO tickers (
                ticker, company_name, sector, industry, exchange, address, official_site, 
                revenue_ttm, market_cap, fiscal_year_end, last_price, updated_at
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
        `, tickerSymbol, data["Name"], data["Sector"], data["Industry"], data["Exchange"], 
            data["Address"], data["OfficialSite"], data["RevenueTTM"], data["MarketCapitalization"], 
            data["FiscalYearEnd"], lastPrice)
        if err != nil {
            log.Fatal(err)
        }

    } else if action == "update" {

        // Update data in database
        _, err = db.Exec(`
            UPDATE tickers SET
                last_price = ?, updated_at = CURRENT_TIMESTAMP
            WHERE ticker = ?
        `, lastPrice, tickerSymbol)
        if err != nil {
            log.Fatal(err)
        }
    }

    /*
    fmt.Println("Data Map:")
    for key, value := range data {
        fmt.Println(key, value)
    }
    */

    fmt.Printf("\n   Exchange: %s\n", data["Exchange"])
    fmt.Printf("   Sector: %s\n", data["Sector"])
    fmt.Printf("   Industry: %s\n", data["Industry"])
    fmt.Printf("   Fiscal Year End: %s\n", data["FiscalYearEnd"])
    fmt.Printf("   Latest Quarter: %s\n", data["LatestQuarter"])

    fmt.Printf("\n   Address: %s\n", data["Address"])

    fmt.Printf("\n   Official Website: %s\n", data["OfficialSite"])

    revenueTTM := formatRevenueTTM(data["RevenueTTM"])
    fmt.Printf("\n   Market Cap (B): %s\n", formatRevenueTTM(data["MarketCapitalization"]))
    fmt.Printf("   Revenue TTM (B): %s\n", revenueTTM)
    fmt.Printf("   Dividend Date: %s\n", data["DividendDate"])

    fmt.Printf("\n   52 Week High: %s\n", data["52WeekHigh"])
    fmt.Printf("   52 Week Low: %s\n", data["52WeekLow"])
    fmt.Printf("   Analyst Target Price: %s\n", data["AnalystTargetPrice"])

    fmt.Printf("\n   PE Ratio: %s\n", data["PERatio"])
    fmt.Printf("   Beta: %s\n", data["Beta"])
    fmt.Printf("   Forward PE: %s\n", data["ForwardPE"])
    fmt.Printf("   Trailing PE: %s\n", data["TrailingPE"])
}



func formatRevenueTTM(revenueTTM string) string {
    revenueTTMFloat, _ := strconv.ParseFloat(revenueTTM, 64)
    return fmt.Sprintf("%.2fB", revenueTTMFloat/1e9)
}

// getStockQuote sends a request to the Alpha Vantage API to get a stock quote for a ticker symbol.
// It takes the database connection and an optional ticker symbol as arguments.
func getStockQuote(db *sql.DB, tickerSymbol string, action string) {
    apiKey := os.Getenv("ALPHAVANTAGE_API_KEY")
    if apiKey == "" {
        fmt.Println("ALPHAVANTAGE_API_KEY environment variable is not set.")
        return
    }

    if tickerSymbol == "" {

        reader := bufio.NewReader(os.Stdin)

        fmt.Print("\nEnter a ticker symbol: (e.g., AAPL, GOOG) [Ctrl+D to cancel] ")

        input, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                fmt.Println("Cancelled")
                fmt.Println()
                return
            }
            fmt.Println("Error reading input:", err)
            return
        }

        tickerSymbol = strings.TrimSpace(input)

    }

    url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", tickerSymbol, apiKey)
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println(err)
        return
    }
    
    defer resp.Body.Close()

    /*
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println(string(body))

    if info, ok := data["Information"]; ok {
        fmt.Println(info)
        return
    }

    */

    var data map[string]interface{}

    err = json.NewDecoder(resp.Body).Decode(&data)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Check for quota exceeded message
    if _, ok := data["Information"]; ok {
        fmt.Println("Daily API quota exceeded. Please refer to Alpha Vantage's premium plans for higher limits.")
        fmt.Println()
        return
    }

    quote := data["Global Quote"].(map[string]interface{})

    if quote["01. symbol"] == "" {
        fmt.Println("Invalid ticker symbol:", tickerSymbol)
        fmt.Println()
        return
    }

    fmt.Printf("Symbol: %s Price: %s Open: %s Change: %s Change Percent: %s\n", quote["01. symbol"], quote["05. price"], quote["02. open"], quote["09. change"], quote["10. change percent"])
    fmt.Printf("   High: %s Low: %s Previous Close: %s\n", quote["03. high"], quote["04. low"], quote["08. previous close"])

    getStockOverview(db,quote["01. symbol"],quote["05. price"],action)

}

// reuseTicker allows the user to choose a previously entered ticker symbol from the database.
func reuseTicker(db *sql.DB) {
    // Retrieve unique ticker symbols from the database
    rows, err := db.Query("SELECT id, company_name, ticker, exchange, last_price, updated_at FROM tickers GROUP BY ticker")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    var tickers []Ticker 

    for rows.Next() {
        var ticker Ticker
        var updatedAt interface{}
        err := rows.Scan(&ticker.Id, &ticker.CompanyName, &ticker.Ticker, &ticker.Exchange, &ticker.LastPrice, &updatedAt)
        if err != nil {
            log.Fatal(err)
        }
        if updatedAt != nil {
            t := updatedAt.(time.Time)
            ticker.UpdatedAt = t.Format("2006-01-02T15:04:05")
        } else {
            ticker.UpdatedAt = ""
        }
        tickers = append(tickers, ticker)
    }

    if len(tickers) == 0 {
        fmt.Println("No stock stickers found")
        fmt.Println()
        return
    }

    fmt.Println("Previous ticker symbols")
    fmt.Println()

    // Assign numbers to each result and ask the user to choose a ticker symbol
    for i, ticker := range tickers {
        var updatedat string
        if ticker.UpdatedAt != "" {
             updatedat = extractDate(ticker.UpdatedAt) + " at " + formatTime(ticker.UpdatedAt) 
        }
        fmt.Printf("%d. %s (%s:%s) %s on %s\n", i+1, ticker.CompanyName, ticker.Ticker, ticker.Exchange, ticker.LastPrice, updatedat)
    }

    fmt.Printf("\nEnter the row number (%d-%d): ", 1, len(tickers))
    var choice int
    _, err = fmt.Scan(&choice)
    if err != nil {
        fmt.Println("Invalid input")
        return
    }

    // Validate user input
    if choice < 1 || choice > len(tickers) {
        fmt.Println("Invalid choice")
        return
    }

    fmt.Println("\n1. Reuse")
    fmt.Println("2. Delete")
    fmt.Println("3. Return to previous menu")
    fmt.Println()
    fmt.Print("Enter your choice: ")
    var action int
    _, err = fmt.Scan(&action)
    if err != nil {
        fmt.Println("Invalid input")
        return
    }

    switch action {
    case 1:
        // Get stock overview for chosen ticker symbol
        chosenTicker := tickers[choice-1].Ticker
        getStockQuote(db, chosenTicker,"update")
    case 2:
        // Delete the selected ticker symbol
        deleteTicker(db, tickers[choice-1].Id)
    case 3:
        // Return to previous menu
        return
    default:
        fmt.Println("Invalid choice")
    }

}

// getTickerMenu prompts the user to enter a new ticker symbol or reuse a previous one.
func tickerMenu(db *sql.DB) {
    fmt.Println("\nTicker menu:")
    fmt.Println()
    fmt.Println("1. Enter a new ticker symbol")
    fmt.Println("2. Re-use/delete a previous ticker symbol")

    fmt.Println()

    var option int
    fmt.Scanln(&option)

    fmt.Println()

    switch option {
    case 1:
        // Enter a new ticker symbol
        getStockQuote(db,"","insert")
    case 2:
        // Re-use a previous ticker symbol
        reuseTicker(db)
    }
}

// getLatestRecords returns the latest TreasuryData records by security description.
func getLatestRecords(data []TreasuryData) map[string]TreasuryData {
    latestRecords := make(map[string]TreasuryData)
    for _, treasury := range data {
            recordDate, err := time.Parse("2006-01-02", treasury.RecordDate)
            if err != nil {
                    // Handle error, e.g., log or return an error
                    continue
            }
            if existing, ok := latestRecords[treasury.SecurityDesc]; !ok || recordDate.After(existingRecordDate(existing)) {
                    latestRecords[treasury.SecurityDesc] = treasury
            }
    }
    return latestRecords
}

// existingRecordDate returns the record date of an existing TreasuryData record.
func existingRecordDate(existing TreasuryData) time.Time {
    recordDate, err := time.Parse("2006-01-02", existing.RecordDate)
    if err != nil {
            // Handle error, e.g., log or return an error
            // You might consider returning a zero time or a specific error
            return time.Time{}
    }
    return recordDate
}

// getTreasury sends a request to the Treasury API to get the latest treasury avg bond, note, bill data.
// and calculates the spread between them.
func getTreasury() {

	// Construct the API request
    // sorted by record date in descending order since it goes back years and we want the latest data
	url := fmt.Sprintf("https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v2/accounting/od/avg_interest_rates?sort=-record_date&format=json")

	// Send the request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}

    //println(resp.Status)
    //println(resp.Body)

	// Read the response body
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

    //fmt.Println(string(body))

	// Unmarshal the JSON response

    var response TreasuryResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
		return
	}
    var tBill, tNote, tBond float64

	// Print the coordinates
	if len(response.Data) > 0 {
        println("\nLatest U.S. Treasury Avg Interest Rates:\n")
        latestRecords := getLatestRecords(response.Data)
        // Access the latest records by security description
        for securityDesc, latestRecord := range latestRecords {
            //fmt.Printf("Security Desc: %s, Record Date: %s\n", securityDesc, latestRecord.RecordDate)
            if securityDesc == "Treasury Bills" {
                fmt.Printf("Bills: %s\n", latestRecord.AvgInterestRateAmt)
                tBill, _ = strconv.ParseFloat(latestRecord.AvgInterestRateAmt, 64)
            }
            if securityDesc == "Treasury Notes" {
                fmt.Printf("Notes: %s\n", latestRecord.AvgInterestRateAmt)
                tNote, _ = strconv.ParseFloat(latestRecord.AvgInterestRateAmt, 64)
            }
            if securityDesc == "Treasury Bonds" {
                fmt.Printf("Bonds: %s\n", latestRecord.AvgInterestRateAmt)
                tBond, _ = strconv.ParseFloat(latestRecord.AvgInterestRateAmt, 64)
            }
        }
        fmt.Printf("\nSpread (Bond to Bill): %.2f\n", tBond - tBill)
        fmt.Printf("Spread (Note to Bill): %.2f\n", tNote - tBill)
        fmt.Printf("Spread (Bond to Note): %.2f\n", tBond - tNote)
        fmt.Println()
    }

}



// main is the entry point of the polyapi CLI tool.
//
//

func main() {

    db := createDB()

    fmt.Println()
	fmt.Println("polyAPI CLI")
	fmt.Println("-----------")
    fmt.Println()

	// Main menu
	for {
		fmt.Println("Main Menu:")
        fmt.Println()
		fmt.Println("1. Get weather for an address")
        fmt.Println("2. Get stock quote")
        fmt.Println("3. Get treasury data")
		fmt.Println("4. Exit")
        fmt.Println()

		var option string
		fmt.Print("Enter your option: ")
		fmt.Scanln(&option)

		switch option {
		case "1":
			
            geocodeMenu(db)
        case "2":
            tickerMenu(db)
        case "3":
            getTreasury()
		case "4":
            defer db.Close()
			fmt.Println("\nExiting...")
			return
		default:
			fmt.Println("\nInvalid option")
		}
	}

}
