package main

/**
query location from google maps
*/

import (
  "bufio"
  "encoding/json"
  "flag"
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "os"
)

const GMAP_BASE = "https://maps.google.com/maps/api/geocode/json?sensor=false&language=zh-CN&address="

type LatLng struct {
  Lat float64
  Lng float64
}

type AddressComponent struct {
  LongName  string   `json:"long_name"`
  ShortName string   `json:"short_name"`
  Types     []string `json:"types"`
}

type Result struct {
  AddressComponents []AddressComponent
  FormattedAddress  string `json:"formatted_address"`
  Geometry          struct {
    Location     LatLng `json:"location"`
    LocationType string `json:"location_type"`
    Viewport     struct {
      NorthEast LatLng `json:"northeast"`
      SouthWest LatLng `json:"southwest"`
    } `json:"viewport"`
  }
  Types []string `json:"types"`
}

type GMap struct {
  Results []Result
  Status  string `json:"status"`
}

func show_usage() {
  fmt.Fprintf(os.Stderr,
    "Usage: %s \n",
    os.Args[0])
  flag.PrintDefaults()
}

var (
  input_file  string
  output_file string
  verbose     bool //  Make the operation more talkative
)

func main() {
  var (
    inf, ouf *os.File
    err      error
    outline  string
    gmaps    GMap
    line     string
  )

  flag.Usage = show_usage
  flag.StringVar(&input_file, "i", "", `address file,split by '\n' `)
  flag.StringVar(&output_file, "o", "", `output file`)
  flag.BoolVar(&verbose, "v", false, "Make the operation more talkative")
  flag.Parse()

  if len(input_file) == 0 || len(output_file) == 0 {
    fmt.Println("no input file or no output file.")
    os.Exit(1)
  }

  inf, err = os.Open(input_file)
  defer inf.Close()

  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  ouf, err = os.OpenFile(output_file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModeAppend|0666)
  defer ouf.Close()

  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  r := bufio.NewReader(inf)
  for {
    line, err = Readline(r)
    if err != nil {
      break
    }
    if len(line) == 0 {
      continue
    }
    if verbose {
      fmt.Printf("Query: %s\n", line)
    }
    gmaps, err = AddressToLocation(line)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Query %s has error %v\n", line, err)
      continue
    }
    if gmaps.Status == "OK" && len(gmaps.Results) > 0 {
      outline = fmt.Sprintf("\"%s\",%f,%f\n", line, gmaps.Results[0].Geometry.Location.Lat, gmaps.Results[0].Geometry.Location.Lng)
    } else {
      outline = fmt.Sprintf("\"%s\",0.00,0.00\n", line)
    }
    ouf.WriteString(outline)
  }
}

func Readline(r *bufio.Reader) (string, error) {
  var (
    isPrefix bool  = true
    err      error = nil
    line, ln []byte
  )
  for isPrefix && err == nil {
    line, isPrefix, err = r.ReadLine()
    ln = append(ln, line...)
  }
  return string(ln), err
}

func AddressToLocation(address string) (gmaps GMap, err error) {
  var (
    resp *http.Response
    body []byte
  )
  query := url.QueryEscape(address)
  if resp, err = http.Get(fmt.Sprintf("%s%s", GMAP_BASE, query)); err != nil {
    return
  }
  defer resp.Body.Close()
  if body, err = ioutil.ReadAll(resp.Body); err != nil {
    return
  }
  err = json.Unmarshal(body, &gmaps)
  return
}
