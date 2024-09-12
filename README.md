# An API app written in Go

polyapi retrieves information from many (poly) disparate APIs within a CLI app that you may have historically retrieved through web pages. Go is a cross-platform, statically typed and compiled programming language. A dev container (development container) is a self-contained dev environment with a Docker container allowing portability and instant creation of a dev environment without manual installation.

## Lastest update

This is a new project and actively under development.

## Currently implemented functionality
1. Reads environment variables like API keys
1. Shows weather forecasts and observations from the nearest weather stations by geocoding an address entered
1. Shows stock ticker data
1. Stores validated addresses and ticker symbols in a local SQLite3 database for re-use or deletion
1. Also stores the last temperature and last ticker price on each API call with an `updated_at` timestamp
1. Retrieves latest average US Treasury bond, note and bill rates and app calculates spreads. 
1. Retrieves latest PPI and CPI from U.S. Bureau of Labor Statistics
1. Retrieves additional economic data like GDP and unemployment rate from U.S. Federal Reserve
1. Query your Salesforce.com instance for contacts
1. Show weekly schedules for NFL and College football, and scores if game underday and links to roster, stats


![screenshot of main menu and retrieving weather](./docs/images/polyapi-address-weather.png)

## SQLite3 local database

At program start, a db directory and `polyapi.db` are created. `db/polyapi.db` is added to a `.gitignore` file so it will not be included in the code repository.

## dev container

I'm using a dev container so I don't have to install Go on my Mac. All I need a is a Docker daemon, which in my case is `colima` and VS Code with the dev container extension.

## API providers

### ESPN football games this week

ESPN has NFL and College Football endpoints. If the game is in progress or over, the app presents score, a headline and a video replay link. If a game (event) is selected, additional links are shown for game, roster, stats

### Weather for Addresses (USGOV)

NOAA's API provides weather forecast information but requires latitude and longitude coordinates. The U.S. Census bureau has a geocoding API that returns coordinates based on a valid address. No API keys required for both APIs.

### Stock Quotes (Alpha Vantage)

Get a [free API key](https://www.alphavantage.co/support/#api-key) to retrieve stock quote data. Add an environment variable in your configuration script e.g., `.zshrc` or `bashrc` that the dev container reads. 

```sh
export ALPHAVANTAGE_API_KEY=""
```

### US Treasury Rates (USGOV)

The U.S. Treasury has a [public API](https://fiscaldata.treasury.gov/api-documentation/) to retrieve financial data including their [rate API](https://fiscaldata.treasury.gov/datasets/average-interest-rates-treasury-securities/average-interest-rates-on-u-s-treasury-securities#api-quick-guide) for [average treasury rates](https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v2/accounting/od/avg_interest_rates?sort=-record_date).  No API key required.

### US PPI and CPI (USGOV)

The U.S. Bureau of Labor Statistics has a [public API](https://www.bls.gov/developers/api_faqs.htm) to retrieve the producer price index "PPI" and consumer price index "CPI"

### US Federal Reserve (USGOV)

The U.S. Federal Reserve has a [public API called FRED](https://www.bls.gov/developers/home.htm#) to retrieve economic statistics. It does require an [API key](https://fred.stlouisfed.org/docs/api/api_key.html) Add an environment variable in your configuration script e.g., `.zshrc` or `bashrc` that the dev container reads. 

```sh
export FRED_API_KEY=""
```

### Query contacts (Salesforce.com)

#### Currently implemented functionality
1. Reads environment variables required (Consumer Key, Consumer Secret, Salesforce Url)
1. Retrieves an OAuth Access Token from Salesforce 
1. Lets users search for Contacts (by contact first & last name, account name, email address) and prints results
1. Retry logic to get a new OAuth Access Token if a Salesforce call fails e.g., token has expired

Register for [a complimentary developer account](https://developer.salesforce.com/signup) to use with Salesforce API testing

#### Authentication

Instance URL, Consumer Key and Consumer Secret are read as environment variables which you place in `.zshrc` or `.bashrc`

```sh
# set SalesForce environment variables
export SALESFORCE_CONSUMER_KEY_1=""
export SALESFORCE_CONSUMER_SECRET_1=""
export SALESFORCE_URL_1=""
```

Retrieve the Url from the Salesforce UI, View Profile and the Url is under your profile user name.
 
Retrieve the Consumer Key and Consumer Secret from the Salesforce UI, View Setup, App Manager, Connected Apps.

The app authenticates uses these environment variables and generates an OAuth Access Token that is used for SOQL Salesforce calls.

### ESPN sports scores and schedules

Retrieves shedules, game day details, and live scores of NFL and College football games

## Creating a binary

`.gitignore` is set to ignore `main` and `polyapi` binaries to reduce repository size. `git build .` uses the OS and architecture of the development machine. To create other binaries, use examples like:

```sh
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o polyapi_darwin_arm64
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o polyapi_linux_amd64
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o polyapi_windows_amd64.exe
```

## Resources

[Go](https://go.dev/)

[Dev Container specification](https://containers.dev/implementors/spec/)

## License

This project is licensed under the [MIT License](LICENSE)

## Contributing

### Disclaimer: Unmaintained and Untested Code

Please note that this program is not actively maintained or tested. While it may work as intended, it's possible that it will break or behave unexpectedly due to changes in dependencies, environments, or other factors.

Use this program at your own risk, and be aware that:
1. Bugs may not be fixed
1. Compatibility issues may arise
1. Security vulnerabilities may exist

If you encounter any issues or have concerns, feel free to open an issue or submit a pull request.