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
    "io"
    "database/sql"
    "sort"
    "time"
    "bytes"
    "net/url"
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

type Measurement struct {
	UnitCode      string  `json:"unitCode"`
	Value         *float64 `json:"value"` 
	QualityControl string  `json:"qualityControl"`
}

type CloudLayer struct {
	Base   Measurement `json:"base"`
	Amount string      `json:"amount"`
}

type Properties struct {
	ID                       string           `json:"@id"`
	Type                     string           `json:"@type"`
	Elevation                Measurement      `json:"elevation"`
	Station                  string           `json:"station"`
	Timestamp                string           `json:"timestamp"`
	RawMessage               string           `json:"rawMessage"`
	TextDescription          string           `json:"textDescription"`
	Icon                     string           `json:"icon"`
	PresentWeather           []interface{}    `json:"presentWeather"`
	Temperature              Measurement      `json:"temperature"`
	Dewpoint                 Measurement      `json:"dewpoint"`
	WindDirection            Measurement      `json:"windDirection"`
	WindSpeed                Measurement      `json:"windSpeed"`
	WindGust                 Measurement      `json:"windGust"`
	BarometricPressure       Measurement      `json:"barometricPressure"`
	SeaLevelPressure         Measurement      `json:"seaLevelPressure"`
	Visibility               Measurement      `json:"visibility"`
	MaxTemperatureLast24Hours Measurement     `json:"maxTemperatureLast24Hours"`
	MinTemperatureLast24Hours Measurement     `json:"minTemperatureLast24Hours"`
	PrecipitationLastHour    Measurement      `json:"precipitationLastHour"`
	PrecipitationLast3Hours  Measurement      `json:"precipitationLast3Hours"`
	PrecipitationLast6Hours  Measurement      `json:"precipitationLast6Hours"`
	RelativeHumidity         Measurement      `json:"relativeHumidity"`
	WindChill                Measurement      `json:"windChill"`
	HeatIndex                Measurement      `json:"heatIndex"`
	CloudLayers              []CloudLayer     `json:"cloudLayers"`
}

type ObservationResponse struct {
	Properties Properties   `json:"properties"`
}

type Geometry struct {
    Type        string    `json:"type"`
    Coordinates []float64 `json:"coordinates"`
}

type Station struct {
    Geometry Geometry `json:"geometry"`
    Properties struct {
        Name              string `json:"name"`
        StationIdentifier string `json:"stationIdentifier"`
    } `json:"properties"`
}

type StationsResponse struct {
    Features []Station `json:"features"`
}

type Observation struct {
    Properties Properties `json:"properties"`
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

type BLSRequest struct {
    SeriesID  []string `json:"seriesid"`
    StartYear string   `json:"startyear"`
    EndYear   string   `json:"endyear"`
}

type BLSResponse struct {
    Results struct {
        Series []BLSSeries `json:"series"`
    } `json:"Results"`
}

type BLSSeries struct {
    SeriesID string     `json:"seriesID"`
    Data     []BLSEntry `json:"data"`
}

type BLSEntry struct {
    Year     string `json:"year"`
    Period   string `json:"period"`
    Value    string `json:"value"`
    Footnote []struct {
        Text string `json:"text"`
    } `json:"footnotes"`
}

type FredResponse struct {
	Observations []struct {
		Date  string `json:"date"`
		Value string `json:"value"`
	} `json:"observations"`
}

type Event struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    ShortName string    `json:"shortName"`
    Competitions []struct {
        Competitors []struct {
            Team struct {
                ID   string `json:"id"`
                Name string `json:"name"`
                DisplayName string `json:"displayName"`
                Links []struct {
                    Href string `json:"href"`
                    Text string `json:"text"`
                } `json:"links"`
            } `json:"team"`
            HomeAway string `json:"homeAway"`
			Links []struct {
				Href string `json:"href"`
				Text string `json:"text"`
			} `json:"links"`
            Score string `json:"score"`
            Records []struct {
                Name  string `json:"name"`
                Summary string `json:"summary"`
            } `json:"records"`
        } `json:"competitors"`
        Broadcasts []struct {
            Market string `json:"market"`
            Names[]string `json:"names"`
        } `json:"broadcasts"`
        Headlines []Headline `json:"headlines"`
    } `json:"competitions"`
    Status struct {
        DisplayClock string `json:"displayClock"`
        Period int `json:"period"`
        Type struct {
            Detail string `json:"detail"`
            ShortDetail string `json:"shortDetail"`
            Description string `json:"description"`
            State string `json:"state"`
            Completed bool `json:"completed"`
        } `json:"type"`
    } `json:"status"`
    Links []struct {
        Href string `json:"href"`
        Text string `json:"text"`
    } `json:"links"`
    Weather struct {
        DisplayValue string `json:"displayValue"`
        Temperature int `json:"temperature"`
        Link struct {
            Href string `json:"href"`
            Text string `json:"text"`
        } `json:"link"`
    } `json:"weather"`
}

type Headline struct {
    Type string `json:"type"`
    Description string `json:"description"`
    ShortLinkText string `json:"shortLinkText"`
    Video []struct {
        Links struct {
            Web struct {
                Href string `json:"href"`
            } `json:"web"`
        } `json:"links"`
    } `json:"video"`
}

type NFLSchedule struct {
    Events []Event `json:"events"`
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

const salesforceAPIBaseURL = "/services/data/v54.0"

type Salesforce struct {
	Url string
	ConsumerKey string
	ConsumerSecret string
	AccessToken string
}

// Define a generic interface to handle different Salesforce objects
type SalesforceObject interface{}

// Define a Contact struct
type Contact struct {
    Id        	string `json:"Id"`
    FirstName 	string `json:"FirstName"`
    LastName  	string `json:"LastName"`
	Account 	Account `json:"Account"`
    Email     	string `json:"Email"`
    Phone     	string `json:"Phone"`
	Description string `json:"Description"`
}

// Define an Account struct
type Account struct {
    Id          string `json:"Id"`
    Name        string `json:"Name"`
    Type        string `json:"Type"`
    Description string `json:"Description"`
    Website     string `json:"Website"`
    Industry    string `json:"Industry"`
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

// extractDate extracts the date from a timestamp, handling both with and without timezone offset.
func extractDate(timestamp string) string {
    if timestamp == "" {
        return "Unknown Date"
    }

    // Define formats for timestamps with and without timezone offset
    formats := []string{
        time.RFC3339,         // e.g., "2024-08-26T14:25:00+00:00"
        "2006-01-02T15:04:05", // e.g., "2024-08-26T14:25:00" (without offset)
    }

    var t time.Time
    var err error

    // Try parsing with each format
    for _, format := range formats {
        t, err = time.Parse(format, timestamp)
        if err == nil {
            return t.Format("2006-01-02") // Successfully parsed, return formatted date
        }
    }

    // If none of the formats worked
    fmt.Println("Error parsing timestamp:", err)
    return "Unknown Date"
}


// formatTime formats the time part of the timestamp.
func formatTime(timestamp string) string {
    if timestamp == "" {
        return "Unknown Time"
    }

    // Define formats for timestamps with and without timezone offset
    formats := []string{
        time.RFC3339,         // e.g., "2024-08-26T14:25:00+00:00"
        "2006-01-02T15:04:05", // e.g., "2024-08-26T14:25:00" (without offset)
    }

    var t time.Time
    var err error

    // Try parsing with each format
    for _, format := range formats {
        t, err = time.Parse(format, timestamp)
        if err == nil {
            return t.Format("03:04 PM") // Format time as "03:04 PM"
        }
    }

    // If none of the formats worked
    fmt.Println("Error parsing timestamp:", err)
    return "Unknown Time"
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

//
func generateGoogleMapsURL(lat, lon string) string {
    return fmt.Sprintf("https://www.google.com/maps/search/?api=1&query=%s,%s", lat, lon)
}

// Helper function to convert Celsius to Fahrenheit
func celsiusToFahrenheit(celsius float64) float64 {
    return (celsius * 9 / 5) + 32
}

// Fetches the nearest observation stations and returns their information
func getNearestStations(lat, lon string) ([]Station, error) {
    url := fmt.Sprintf("https://api.weather.gov/points/%s,%s/stations", lat, lon)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var stationsResponse StationsResponse
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    err = json.Unmarshal(body, &stationsResponse)
    if err != nil {
        return nil, err
    }

    // Assuming you want the 2 closest stations
    closestStations := stationsResponse.Features[:4]
    return closestStations, nil
}

// Fetches the observation data for a specific station
func getObservation(stationID string) (Observation, error) {
    url := fmt.Sprintf("https://api.weather.gov/stations/%s/observations/latest", stationID)
    resp, err := http.Get(url)
    if err != nil {
        return Observation{}, err
    }
    defer resp.Body.Close()

    var observation Observation
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return Observation{}, err
    }

    err = json.Unmarshal(body, &observation)
    if err != nil {
        return Observation{}, err
    }

    return observation, nil
}

// Print observation information
func printObservation(observation Properties) {

    var observationDateTime = extractDate(observation.Timestamp) + " at " + formatTime(observation.Timestamp)

	fmt.Printf("  Timestamp: %s\n", observationDateTime)
    if observation.Temperature.Value != nil {
	    fmt.Printf("  Temperature: %.2f°F\n", celsiusToFahrenheit(*observation.Temperature.Value))
    }
    if observation.Dewpoint.Value != nil {
	    fmt.Printf("  Dewpoint: %.2f°F\n", celsiusToFahrenheit(*observation.Dewpoint.Value))
    }
    if observation.WindSpeed.Value != nil {
        fmt.Printf("  Wind Speed: %.2f km/h\n", *observation.WindSpeed.Value)
    }
	if observation.WindDirection.Value != nil {
        fmt.Printf("  Wind Direction: %.2f°\n", *observation.WindDirection.Value)
    }
    if observation.RelativeHumidity.Value != nil {
        fmt.Printf("  Humidity: %.2f%%\n", *observation.RelativeHumidity.Value)
    }
	if observation.BarometricPressure.Value != nil {
        fmt.Printf("  Pressure: %.2f Pa\n", *observation.BarometricPressure.Value)
    }
	if observation.Visibility.Value != nil {
        fmt.Printf("  Description: %s\n", observation.TextDescription)
    }
    fmt.Println()
}


// getNOAAWeather sends a request to the NOAA API to get the weather forecast for a location.
// it takes the latitude and longitude of the location as arguments.
// it is called from geocode function.

func getNOAAWeather(lat, lon string, db *sql.DB, addressId int) {
    reader := bufio.NewReader(os.Stdin)
    
    // Fetch nearest stations
    stations, err := getNearestStations(lat, lon)
    if err != nil {
        fmt.Println("Error fetching nearest stations:", err)
        return
    }

    // Print observation data for the 2 closest stations
    println("\nNOAA weather stations: (sorted by nearest to farthest)")
    println()
    for i, station := range stations {

        lat := station.Geometry.Coordinates[1]
        lon := station.Geometry.Coordinates[0]
        mapsURL := generateGoogleMapsURL(fmt.Sprintf("%f", lat), fmt.Sprintf("%f", lon))

        fmt.Printf("Station %d: %s\n", i+1, station.Properties.Name)
        fmt.Printf("  Identifier: %s\n", station.Properties.StationIdentifier)
        fmt.Printf("  Location: %s\n", mapsURL)
        observation, err := getObservation(station.Properties.StationIdentifier)
        if err != nil {
            fmt.Println("Error fetching observation data:", err)
            continue
        }
        printObservation(observation.Properties)
    }

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

    println()
    googleMapsURL := generateGoogleMapsURL(lat, lon)
    fmt.Printf("%s\n", googleMapsURL)

    // Submenu
    for {
        fmt.Println("\nNOAA Weather Submenu:")
        fmt.Println()
        fmt.Println("1. Forecast")
        fmt.Println("2. Hourly Forecast")
        fmt.Println("3. Choose another address")
        fmt.Println("4. Main Menu")
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
            geocodeMenu(db)
        case "4":
            reader.Reset(os.Stdin)
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

        googleMapsURL := generateGoogleMapsURL(fmt.Sprintf("%f", response.Result.AddressMatches[0].Coordinates.Y), fmt.Sprintf("%f", response.Result.AddressMatches[0].Coordinates.X))
        fmt.Printf("  %s\n", googleMapsURL)

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
        fmt.Printf("%d. %s ~ %s\n", i+1, address.MatchedAddress, extraInfo)
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
    fmt.Println()
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

func processBLSData(series BLSSeries) {

    switch series.SeriesID {
    case "PCU22112222112241":
        fmt.Println("\nProducer Price Index (PPI) Data:")
    case "CUUR0000SA0L1E":
        fmt.Println("\nConsumer Price Index (CPI) Data, less food & energy:")
    case "CUSR0000SA0":
        fmt.Println("\nConsumer Price Index (CPI) Data:")
    case "LNS14000000":
        fmt.Println("\nUnemployment Rate Data:")
    case "CES0000000001":
        fmt.Println("\nNonfarm Payroll Data:")
    default:
        fmt.Println("\n")
    }

    fmt.Printf("Series: %s\n", series.SeriesID)

    // Sort data by year and period
    sort.Slice(series.Data, func(i, j int) bool {
        yearI, _ := strconv.Atoi(series.Data[i].Year)
        yearJ, _ := strconv.Atoi(series.Data[j].Year)
        return yearI < yearJ || (yearI == yearJ && series.Data[i].Period < series.Data[j].Period)
    })

    latestMonth := series.Data[len(series.Data)-1]
    previousMonth := series.Data[len(series.Data)-2]

    latestValue, _ := strconv.ParseFloat(latestMonth.Value, 64)
    previousValue, _ := strconv.ParseFloat(previousMonth.Value, 64)
    change := latestValue - previousValue
    percentageChange := (change / previousValue) * 100

    fmt.Printf("Latest month: %s - Value: %f\n", latestMonth.Period, latestValue)
    fmt.Printf("Previous month: %s - Value: %f\n", previousMonth.Period, previousValue)
    fmt.Printf("Change: %.2f (%.2f%%)\n", change, percentageChange)

    // Calculate 12-month change if there's enough data
    if len(series.Data) >= 13 {
        sameMonth12MonthsAgo := series.Data[len(series.Data)-13]
        sameMonth12MonthsAgoValue, _ := strconv.ParseFloat(sameMonth12MonthsAgo.Value, 64)
        twelveMonthChange := latestValue - sameMonth12MonthsAgoValue
        twelveMonthPercentageChange := (twelveMonthChange / sameMonth12MonthsAgoValue) * 100
        fmt.Printf("12-month change: %.2f (%.2f%%)\n", twelveMonthChange, twelveMonthPercentageChange)
    } else {
        fmt.Println("Not enough data to calculate 12-month change")
    }
    fmt.Println()
}


func getBLSData() {

	// Get the current year and the previous year
	currentYear := time.Now().Year()
	previousYear := currentYear - 1

    // Define the data for the POST request
    reqData := BLSRequest{
        SeriesID:  []string{"PCU22112222112241", "CUUR0000SA0L1E", "CUSR0000SA0", "LNS14000000", "CES0000000001"},
        StartYear: strconv.Itoa(previousYear),
        EndYear:   strconv.Itoa(currentYear),
    }

    // Marshal the request data into JSON
    jsonData, err := json.Marshal(reqData)
    if err != nil {
        fmt.Println("Error marshaling JSON:", err)
        return
    }

    // Make the POST request
    url := "https://api.bls.gov/publicAPI/v2/timeseries/data/"
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error making POST request:", err)
        return
    }
    defer resp.Body.Close()

    // Read the response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Error reading response body:", err)
        return
    }

    // Unmarshal the JSON response
    var blsResponse BLSResponse
    err = json.Unmarshal(body, &blsResponse)
    if err != nil {
        fmt.Println("Error unmarshaling JSON:", err)
        return
    }

    // Process and print the response data
    for _, series := range blsResponse.Results.Series {
        processBLSData(series)
    }
}


func fetchSeriesData(seriesID, startYear, endYear string) {


    apiKey := os.Getenv("FRED_API_KEY")
    if apiKey == "" {
        fmt.Println("FRED_API_KEY environment variable is not set.")
        return
    }

    url := fmt.Sprintf("https://api.stlouisfed.org/fred/series/observations?series_id=%s&observation_start=%s&observation_end=%s&api_key=%s&limit=13&file_type=json&sort_order=desc", seriesID, startYear, endYear, apiKey)
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

    var data FredResponse

    err = json.NewDecoder(resp.Body).Decode(&data)
    if err != nil {
        fmt.Println(err)
        return
    }

	// Print the results with a friendly header based on Series ID

    println()

    fmt.Printf("**")
	switch seriesID {
	case "FEDFUNDS":
		fmt.Printf("Federal Funds Rate")
	case "ICSA":
		fmt.Printf("Initial Claims for Unemployment Insurance")
	case "RSAFS":
		fmt.Printf("Retail Sales")
    case "UNRATE":
        fmt.Printf("Unemployment Rate")
    case "GDP":
        fmt.Printf("Gross Domestic Product")
    case "DGORDER":
        fmt.Printf("Durable Goods Orders")
    case "INDPRO":
        fmt.Printf("Industrial Production")
    case "PCE":
        fmt.Printf("Personal Consumption Expenditures")
	default:
		fmt.Printf("Unknown Series ID")
	}

    fmt.Printf("**")
	fmt.Printf("\nFRED Series ID: %s and %d observations", seriesID, len(data.Observations))
    fmt.Println()

	if len(data.Observations) > 0 {
		latest := data.Observations[0] // The first element is the latest due to descending sort
		fmt.Printf("%s on %s\n", latest.Value, latest.Date)

        // GDP data is quarterly, so calculate quarter-over-quarter change
        if seriesID == "GDP" {
            if len(data.Observations) == 4 {
                // previous quarter calculation
                previousQuarter := data.Observations[1] 
                latestValue, _ := strconv.ParseFloat(latest.Value, 64)
                previousValue, _ := strconv.ParseFloat(previousQuarter.Value, 64)
                quarterChange := latestValue - previousValue
                quarterPercentageChange := (quarterChange / previousValue) * 100
                fmt.Printf("Change from previous quarter: %.2f (%.2f%%) | Value: %s\n", quarterChange, quarterPercentageChange, previousQuarter.Value)

                // previous year calculation
                previousYear := data.Observations[3]
                previousYearValue, _ := strconv.ParseFloat(previousYear.Value, 64)
                yearChange := latestValue - previousYearValue
                yearPercentageChange := (yearChange / previousYearValue) * 100
                fmt.Printf("Change from previous year: %.2f (%.2f%%) | Value: %s\n", yearChange, yearPercentageChange, previousYear.Value)
            } else {
                fmt.Println("Not enough data to calculate quarter-over-quarter and annual change")
            }
            return
        }

		// Ensure there are at least 2 observations to calculate month-over-month change
		if len(data.Observations) > 1 {
			previous := data.Observations[1]
			latestValue, _ := strconv.ParseFloat(latest.Value, 64)
			previousValue, _ := strconv.ParseFloat(previous.Value, 64)
			change := latestValue - previousValue
			percentageChange := (change / previousValue) * 100
			fmt.Printf("Change from previous month (%s): %.2f (%.2f%%) | Value: %s\n", previous.Date, change, percentageChange, previous.Value)
		} else {
			fmt.Println("Not enough data to calculate month-over-month change")
		}

		// Calculate 12-month change if there's enough data
		if len(data.Observations) >= 12 {
			yearAgo := data.Observations[11] // 12th element is data from 12 months ago
            //yearAgo := data.Observations[len(data.Observations)-1] // 12th element is data from 12 months ago
			latestValue, _ := strconv.ParseFloat(latest.Value, 64)
			yearAgoValue, _ := strconv.ParseFloat(yearAgo.Value, 64)
			yearChange := latestValue - yearAgoValue
			yearPercentageChange := (yearChange / yearAgoValue) * 100
			fmt.Printf("Change from 12 months ago: (%s) %.2f (%.2f%%) | Value: %s\n", startYear, yearChange, yearPercentageChange, yearAgo.Value)
		} else {
			fmt.Println("Not enough data to calculate 12-month change")
		}


	} else {
		fmt.Println("No data available in the specified date range")
	}

    println()

}

func getFirstDayOfMonth(date time.Time) time.Time {
    // Construct a new date with the same year and month, but with the day set to 1
    firstDayOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
    return firstDayOfMonth
}

func getFRED() {


    currentDate := time.Now()
    firstDayOfMonth := getFirstDayOfMonth(currentDate)
    oneYearBeforeFirstDayOfMonth := firstDayOfMonth.AddDate(-1, 0, 0)
    oneYearOneMonthBeforeFirstDayOfMonth := oneYearBeforeFirstDayOfMonth.AddDate(0, -1, 0)

    // Convert the dates to strings
    firstDayOfMonthStr := firstDayOfMonth.Format("2006-01-02")
    oneYearBeforeFirstDayOfMonthStr := oneYearBeforeFirstDayOfMonth.Format("2006-01-02")
    oneYearOneMonthBeforeFirstDayOfMonthStr := oneYearOneMonthBeforeFirstDayOfMonth.Format("2006-01-02")

    seriesIDs := []string{"FEDFUNDS", "ICSA", "RSAFS", "UNRATE", "GDP", "PCE"}

    // Fetch data concurrently
    for _, id := range seriesIDs {
        if id == "PCE" {
            fetchSeriesData(id, oneYearOneMonthBeforeFirstDayOfMonthStr, firstDayOfMonthStr)
        } else {
            fetchSeriesData(id, oneYearBeforeFirstDayOfMonthStr, firstDayOfMonthStr)
        }
    }
}

func getFootballSchedule(league string) {

    var url string

    switch league {
    case "NFL":
        url = fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard")
    case "College":
        url = fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/football/college-football/scoreboard")
    }


    resp, err := http.Get(url)
    if err != nil {
        fmt.Println(err)
        return
    }
    
    defer resp.Body.Close()

    var data NFLSchedule

    err = json.NewDecoder(resp.Body).Decode(&data)
    if err != nil {
        fmt.Println(err)
        return
    }

    
    //jsonData, _ := json.MarshalIndent(data, "", "  ")
    //fmt.Println(string(jsonData))

    fmt.Printf("\n%s Schedule:",league)
    fmt.Println()
    fmt.Println()

    for i, event := range data.Events {
        //fmt.Printf("%+v\n", event)
        fmt.Printf("%d. \n", i+1)
        fmt.Printf("%s\n", event.Name)
        fmt.Printf("%s\n", event.Status.Type.Detail)
        if event.Status.Type.State != "post" {
            fmt.Printf("%s\n", strings.Join(event.Competitions[0].Broadcasts[0].Names, ", "))
        }
        fmt.Println()
        if event.Status.Type.State != "pre" && event.Status.Type.State != "post" {
            fmt.Printf("%s period %s\n", event.Status.DisplayClock, event.Status.Period)
        }
        if event.Status.Type.State != "pre" {
            fmt.Printf("%s %s\n", event.Competitions[0].Competitors[0].Team.Name, event.Competitions[0].Competitors[0].Score)
            fmt.Printf("%s %s\n", event.Competitions[0].Competitors[1].Team.Name, event.Competitions[0].Competitors[1].Score)
        }
        
        if event.Status.Type.State == "post" {
            println()
            for _, headline := range event.Competitions[0].Headlines {
                fmt.Println(headline.ShortLinkText)
                fmt.Println(headline.Video[0].Links.Web.Href)
                break
              }

            fmt.Println()
        }

        fmt.Println() 
    }

    fmt.Println()

    for {

        var choice string
        fmt.Print("Enter the number of the event: ('q' to quit) ")
        fmt.Scanln(&choice)

        if choice == "q" {
            break
        } else {
            choiceInt, err := strconv.Atoi(choice)
            
            if err != nil {
                fmt.Println("Invalid input. Please enter a number.")
                continue
            }

            if choiceInt < 1 || choiceInt > len(data.Events) {
                fmt.Println("Invalid input. Event number out of range.")
                continue
            }

            chosenEvent := data.Events[choiceInt-1]

            fmt.Println()

            fmt.Println(chosenEvent.Links[0].Text + ": ", chosenEvent.Links[0].Href)

            if chosenEvent.Status.Type.State == "post" {
                fmt.Println()
                fmt.Println("Expected " + chosenEvent.Weather.Link.Text + ": ", chosenEvent.Weather.DisplayValue, strconv.Itoa(chosenEvent.Weather.Temperature) + "°F")
            }

            fmt.Println()

            fmt.Println("More info: " + chosenEvent.Weather.Link.Href)

            fmt.Println()

            fmt.Println(chosenEvent.Competitions[0].Competitors[0].Team.DisplayName)
            fmt.Println()
            for _, link := range chosenEvent.Competitions[0].Competitors[0].Team.Links {
                fmt.Println(link.Text + ": ", link.Href)
            }

            fmt.Println()

            fmt.Println(chosenEvent.Competitions[0].Competitors[1].Team.DisplayName)
            fmt.Println()
            var repeatCount = 0
            for _, link := range chosenEvent.Competitions[0].Competitors[1].Team.Links {
                if link.Text == "Clubhouse" && repeatCount == 0 {
                    repeatCount++
                    fmt.Println(link.Text + ": ", link.Href)
                }

                if link.Text != "Clubhouse" {
                    fmt.Println(link.Text + ": ", link.Href)
                }
            }

            fmt.Println()
        }

    }

}

func getCollegeFootballSchedule() {

    println("College Football Schedule - TBD")
}

func espnMenu() {

    fmt.Println("\nESPN menu:")
    fmt.Println()
    fmt.Println("1. Next week's NFL schedule")
    fmt.Println("2. Next week's College Football schedule")
    fmt.Println()

    var option int
    fmt.Scanln(&option)

    fmt.Println() 

    switch option {
    case 1:
        getFootballSchedule("NFL")
    case 2:
        getFootballSchedule("College")
    }

}

func setVars(deployment *Salesforce, requiredVars []string) {
    missingVars := []string{}

    for _, varName := range requiredVars {
        value := os.Getenv(varName)
        if value == "" {
            missingVars = append(missingVars, varName)
        } else {
            switch varName {
            case requiredVars[0]:
                deployment.Url = value
            case requiredVars[1]:
                deployment.ConsumerKey = value
            case requiredVars[2]:
                deployment.ConsumerSecret = value
            }
        }
    }

    // If it's not optional and variables are missing, print an error and exit
    if len(missingVars) > 0 {
        fmt.Println("\nError: Missing required environment variables for deployment:\n")
        for _, varName := range missingVars {
            fmt.Printf("  - %s\n", varName)
        }
        os.Exit(1)
    }
}

// isValidDeployment checks if the given Salesforce deployment has valid credentials
func isValidDeployment(s *Salesforce) bool {
    return s.Url != "" && s.ConsumerKey != "" && s.ConsumerSecret != ""
}

func getEnvVars(d1 *Salesforce) Salesforce {

	requiredVars1 := []string{"SALESFORCE_URL_1", "SALESFORCE_CONSUMER_KEY_1", "SALESFORCE_CONSUMER_SECRET_1"}

    // Set the first deployment (required)
    setVars(d1, requiredVars1)

    // Check if credentials are valid and take the first valid deployment
    if isValidDeployment(d1) {
        return *d1
    }

     // If deployment is invalid, return an error
     fmt.Println("Error: Missing required environment variables for both deployments.")
     os.Exit(1)

     return Salesforce{}

}

func getAccessToken(s *Salesforce) (string, error) {

    form := url.Values{}
    form.Add("grant_type", "client_credentials")
    form.Add("client_id", s.ConsumerKey)
    form.Add("client_secret", s.ConsumerSecret)

    // 1. Print request details for debugging
    //fmt.Printf("Sending POST request to: %s\n", s.Url)
    //fmt.Printf("Form data: %v\n", form)

    req, err := http.NewRequest("POST", s.Url+"/services/oauth2/token", strings.NewReader(form.Encode()))
    if err != nil {
        return "", fmt.Errorf("error creating request: %w", err)
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("error making request: %w", err)
    }
    defer resp.Body.Close()


    // 2. Check for successful response status code
    if resp.StatusCode != http.StatusOK {
        defer resp.Body.Close() // Close body even on errors
        body, _ := ioutil.ReadAll(resp.Body)
        return "", fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
    }

    // 3. Print response details for debugging
    // fmt.Printf("Received response with status code: %d\n", resp.StatusCode)



    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("error reading response body: %w", err)
    }
    // fmt.Println("Received response body:")
    // fmt.Println(string(body))

    var result map[string]interface{}
	err = json.Unmarshal(body, &result)
    if err != nil {
        return "", fmt.Errorf("error parsing JSON response: %w, response body: %s", err, string(body))
    }

    accessToken, ok := result["access_token"].(string)
    if !ok {
        return "", fmt.Errorf("couldn't parse access token, response body: %s", string(body))
    }

	s.AccessToken = accessToken

    return accessToken, nil

}

// querySalesforce executes SOQL queries
// 
// Requires:
//		- Salesforce struct with access token
//		- A string with the SOQL query
//		- A destination interface to store the query results
//
func querySalesforce(s *Salesforce, soql string, dest interface{}) error {
    // Create the HTTP request
    req, err := http.NewRequest("GET", s.Url+salesforceAPIBaseURL+"/query?q="+url.QueryEscape(soql), nil)
    if err != nil {
        return fmt.Errorf("error creating request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+s.AccessToken)

    // Make the API call
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("error making request: %w", err)
    }
    defer resp.Body.Close()

    // Check for successful response
    if resp.StatusCode == http.StatusUnauthorized {
        // 401 Unauthorized indicates session expired, try to refresh the token
        fmt.Println("Session expired, refreshing token...\n")

        // Get a new access token
        _, err := getAccessToken(s)
        if err != nil {
            return fmt.Errorf("error refreshing access token: %w", err)
        }

        // Retry the request with the new token
        req.Header.Set("Authorization", "Bearer "+s.AccessToken)
        resp, err = client.Do(req)
        if err != nil {
            return fmt.Errorf("error making request after token refresh: %w", err)
        }
        defer resp.Body.Close()

        // Check the response again after retrying
        if resp.StatusCode != http.StatusOK {
            body, _ := ioutil.ReadAll(resp.Body)
            return fmt.Errorf("unexpected status code after token refresh: %d, response body: %s", resp.StatusCode, string(body))
        }
    } else if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        return fmt.Errorf("unexpected status code: %d, response body: %s", resp.StatusCode, string(body))
    }

    // Parse the JSON response into the provided destination

    /* for debugging json payload 
    // Step 1: Read the entire response body
    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("error reading response body: %w", err)
    }

    // Step 2: Print the raw JSON response
    fmt.Println("Raw JSON response:", string(responseBody))
    */

    err = json.NewDecoder(resp.Body).Decode(dest)
    if err != nil {
        return fmt.Errorf("error parsing JSON response: %w", err)
    }

    return nil
}

func getContacts(salesforce *Salesforce, contactFilter string) ([]Contact, error) {
    soql := fmt.Sprintf("SELECT Id, FirstName, LastName, Email, Account.Name, Phone, Description FROM Contact "+
	"WHERE LastName LIKE '%%%s%%' OR FirstName LIKE '%%%s%%'"+
	"OR Account.Name LIKE '%%%s%%' OR Email LIKE '%%%s%%' ORDER BY LastName",
	contactFilter,contactFilter,contactFilter,contactFilter)
    var contactsResponse struct {
        Records []Contact `json:"records"`
    }
    err := querySalesforce(salesforce, soql, &contactsResponse)
    return contactsResponse.Records, err
}

// FormatCreatedAt takes a date string and returns it formatted as "YYYY-MM-DD HH:MM AM/PM".
func FormatCreatedAt(dateStr string) (string, error) {
    // Parse the input date string (adjust the layout according to your input format)
    createdAt, err := time.Parse("2006-01-02T15:04:05.000-0700", dateStr) // Adjust as necessary
    if err != nil {
        return "", fmt.Errorf("error parsing date: %w", err)
    }

    // Format the date to the desired output
    formattedDate := createdAt.Format("2006-01-02 03:04 PM")
    return formattedDate, nil
}

func printContacts(contacts []Contact) {

    if len(contacts) == 0 {
        fmt.Println("\nNo contacts found.")
        return
    }

    for _, contact := range contacts {
        fmt.Printf("\nContact Name: %s, %s\nAccount: %s\nEmail: %s\nPhone: %s\nDescription:\n\n%s\n\n", contact.LastName, contact.FirstName, contact.Account.Name, contact.Email, contact.Phone, contact.Description)
    }
}

func printObjectCounts(salesforce *Salesforce) {
    // Define a list of SOQL queries for counting different objects
    queries := map[string]string{
        "accounts":      "SELECT COUNT() FROM Account",
        "contacts":      "SELECT COUNT() FROM Contact",
        "opportunities": "SELECT COUNT() FROM Opportunity",
        "tasks":         "SELECT COUNT() FROM Task",
    }

    // Iterate through the queries and print counts
    fmt.Println("\nDeployment counts:\n")
    for object, query := range queries {
        var countResponse struct {
            TotalSize int `json:"totalSize"`
        }
        
        // Execute the query and handle errors
        err := querySalesforce(salesforce, query, &countResponse)
        if err != nil {
            fmt.Printf("Error retrieving count for %s: %s\n", object, err)
            continue // Skip to the next object on error
        }

        // Print the count for the object
        fmt.Printf("  %s: %d\n", object, countResponse.TotalSize)
    }
}

func printSalesforceCreds(s *Salesforce) {

    fmt.Println()
	fmt.Println("Salesforce URL:", s.Url)
	fmt.Println("Salesforce Consumer Key:", s.ConsumerKey)
	fmt.Println("Salesforce Consumer Secret:", s.ConsumerSecret)
	fmt.Println("Generated Salesforce Access Token:", s.AccessToken)
    fmt.Println()

}

func getSalesforce() {

    // Define pointers to Salesforce structs for each deployment and current deployment
    var deployment1, currentDeployment Salesforce

    // Get and print environment variables with Salesforce credentials
    currentDeployment = getEnvVars(&deployment1)

    // Get access token
    _, err := getAccessToken(&currentDeployment)
    if err != nil {
        fmt.Println("Error getting access token:", err)
        return
    }

    printSalesforceCreds(&currentDeployment)
    
    printObjectCounts(&currentDeployment)

    // Check if deployment are valid
    if isValidDeployment(&deployment1) {
        print("\nYou have a valid Salesforce deployment:\n")
        print("\n  ", deployment1.Url)
        print("\n")
    }

    for {
        var contactFilter string
        fmt.Print("\nEnter contact first, last name, email or account name filter (or 'q' to quit): ")
        fmt.Scanln(&contactFilter)

        if contactFilter == "q" {
            break
        }

        contacts, err := getContacts(&currentDeployment, contactFilter)
        if err != nil {
            fmt.Println("Error retrieving contacts:", err)
            continue
        }

        printContacts(contacts)
    }


}

// main is the entry point of the polyapi CLI tool.
//
//

func main() {

    db := createDB()

	// Main menu
	for {
        fmt.Println()
        fmt.Println("polyAPI CLI")
        fmt.Println("-----------")
        fmt.Println()
		fmt.Println("Main Menu:")
        fmt.Println()
		fmt.Println("1. Get weather for an address")
        fmt.Println("2. Get stock quote")
        fmt.Println("3. Get treasury data")
        fmt.Println("4. Get bls economic data like unemployment, ppi, cpi")
        fmt.Println("5. Get federal reserve data like federal funds rate")
        fmt.Println("6. Get ESPN sports data")
        fmt.Println("7. Query Salesforce data")
		fmt.Println("8. Exit")
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
            getBLSData()
        case "5":
            getFRED()
        case "6":
            espnMenu()
        case "7":
            getSalesforce()
		case "8":
            defer db.Close()
			fmt.Println("\nExiting...")
			return
		default:
			fmt.Println("\nInvalid option")
		}
	}

}
