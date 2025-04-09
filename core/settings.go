package core 

import ( 
 "fmt" 
 "log" 
 "os" 
 "strings" 

 "github.com/joho/godotenv" 
) 

var TICKERS []string 

func Load() { 
	err := godotenv.Load(".env")
	if err != nil { 
  log.Fatal("Failed to load environment file") 
 } 
 t := os.Getenv("TICKERS") 
 tickers := strings.Split(t, ",") 
 LoadTikers(tickers) 
 TICKERS = tickers
 fmt.Println(TICKERS) 
}