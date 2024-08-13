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
    "math"
	
)

type GeoCodingResponse struct {
	Result struct {
		AddressMatches []struct {
			Coordinates struct {
				X float64 `json:"x"`
				Y float64 `json:"y"`
			} `json:"coordinates"`
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

// printForecast prints the weather forecast for a location.
// Specifically, it prints the next 4 periods and the last 2 periods.
// A period is a 12-hour time frame.
func printForecast(noaaResponse NOAAWeatherResponse) {

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

// formatTime formats the start time of a period.
func formatTime(startTime string) string {
    startTime = strings.Split(startTime, "T")[1]
    hourMinute := strings.Split(startTime, "-")[0]
    hourMinuteParts := strings.Split(hourMinute, ":")
    hour := hourMinuteParts[0]
    minute := hourMinuteParts[1]
    ampm := "AM"
    if hour >= "12" {
        ampm = "PM"
    }
    if hour == "00" {
        hour = "12"
    } else if hour > "12" {
        hourInt, _ := strconv.Atoi(hour)
        hour = strconv.Itoa(int(math.Mod(float64(hourInt-12), 24)))
    }
    return fmt.Sprintf("%s:%s %s", hour, minute, ampm)
}

// printHourlyForecast prints the hourly weather forecast for a location.
// Specifically, it prints the next 12 hours.
func printHourlyForecast(noaaResponse NOAAWeatherResponse) {
    
    for i := 0; i < 12 && i < len(noaaResponse.Properties.Periods); i++ {
        period := noaaResponse.Properties.Periods[i]
        startTime := formatTime(period.StartTime)
        fmt.Printf("%s %d%s\n", startTime, period.Temperature, period.TemperatureUnit)
        fmt.Printf(" - %s\n", period.ShortForecast)
        fmt.Printf("\n")
    }
}

// getNOAAWeather sends a request to the NOAA API to get the weather forecast for a location.
// it takes the latitude and longitude of the location as arguments.
// it is called from geocode function.

func getNOAAWeather(lat, lon string) {
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
    
    // Submenu
    for {
        fmt.Println("\nNOAA Weather Submenu:\n")
        fmt.Println("1. Forecast")
        fmt.Println("2. Hourly Forecast")
        fmt.Println("3. Return to Main Menu\n")
        
        var option string
        fmt.Print("Enter your option: ")
        option, _ = reader.ReadString('\n')
        option = strings.TrimSpace(option)
        fmt.Println()

        switch option {
        case "1":
            // Call the forecast API
            resp, err := http.Get(noaaResponse.Properties.Forecast)
            if err != nil {
                fmt.Println(err)
                continue
            }
            defer resp.Body.Close()

            body, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                fmt.Println(err)
                continue
            }

            //fmt.Println(string(body))

            err = json.Unmarshal(body, &noaaResponse)
            if err != nil {
                fmt.Println(err)
                return
            }
            
            printForecast(noaaResponse)

        case "2":
            // Call the hourly forecast API
            resp, err := http.Get(noaaResponse.Properties.ForecastHourly)
            if err != nil {
                fmt.Println(err)
                continue
            }
            defer resp.Body.Close()

            body, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                fmt.Println(err)
                continue
            }

            err = json.Unmarshal(body, &noaaResponse)
            if err != nil {
                fmt.Println(err)
                continue
            }

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

func getGeoCode() {

    reader := bufio.NewReader(os.Stdin)

	// Prompt user for address
	fmt.Print("\nEnter address: (must be full address including number & street, city & state, and/or zip code) \n\n")
    address, _ := reader.ReadString('\n')
    address = strings.TrimSpace(address)
    address = strings.ReplaceAll(address, " ", "+")

	// Construct the API request
	url := fmt.Sprintf("https://geocoding.geo.census.gov/geocoder/locations/onelineaddress?address=%s&benchmark=4&format=json", address)

	// Send the request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}

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
        getNOAAWeather(fmt.Sprintf("%f", response.Result.AddressMatches[0].Coordinates.Y), fmt.Sprintf("%f", response.Result.AddressMatches[0].Coordinates.X))
    } else {
        fmt.Println("No coordinates found")
    }
}

// getStockOverview sends a request to the Alpha Vantage API to get an overview of a company
func getStockOverview(tickerSymbol string) {
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
func getStockQuote() {
    apiKey := os.Getenv("ALPHAVANTAGE_API_KEY")
    if apiKey == "" {
        fmt.Println("ALPHAVANTAGE_API_KEY environment variable is not set.")
        return
    }

    fmt.Print("\nEnter a ticker symbol: ")
    var tickerSymbol string
    fmt.Scanln(&tickerSymbol)
    fmt.Println()
    url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", tickerSymbol, apiKey)
    resp, err := http.Get(url)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer resp.Body.Close()

    var data map[string]map[string]string
    err = json.NewDecoder(resp.Body).Decode(&data)
    if err != nil {
        fmt.Println(err)
        return
    }

    quote := data["Global Quote"]

    if quote["01. symbol"] == "" {
        fmt.Println("Invalid ticker symbol:", tickerSymbol)
        return
    }

    fmt.Printf("Symbol: %s Price: %s Open: %s Change: %s Change Percent: %s\n", quote["01. symbol"], quote["05. price"], quote["02. open"], quote["09. change"], quote["10. change percent"])
    fmt.Printf("   High: %s Low: %s Previous Close: %s\n", quote["03. high"], quote["04. low"], quote["08. previous close"])

    getStockOverview(tickerSymbol)

}

// main is the entry point of the polyapi CLI tool.
//
//

func main() {

	fmt.Println("\npolyAPI CLI")
	fmt.Println("-----------\n\n")

	// Main menu
	for {
		fmt.Println("\nMain Menu:\n")
		fmt.Println("1. Return geocode for address & get weather")
        fmt.Println("2. Get stock quote")
		fmt.Println("3. Exit\n")

		var option string
		fmt.Print("Enter your option: ")
		fmt.Scanln(&option)

		switch option {
		case "1":
			
            getGeoCode()
        case "2":
            getStockQuote()
		case "3":
			fmt.Println("\nExiting...")
			return
		default:
			fmt.Println("\nInvalid option")
		}
	}

}
