// nmon2influxdb
// import nmon report in InfluxDB
// author: adejoux@djouxtech.net
package main

import (
	"fmt"
	"github.com/adejoux/influxdbclient"
	"github.com/codegangsta/cli"
	"os"
)

const querytimeformat = "2006-01-02 15:04:05"

func NmonStat(c *cli.Context) {
	// parsing parameters
	config := ParseParameters(c)

	if len(config.Metric) == 0 {
		fmt.Printf("No metric specified ! Use -h option for help !\n")
		os.Exit(1)
	}

	influxdb := influxdbclient.NewInfluxDB(config.InfluxdbServer, config.InfluxdbPort, config.InfluxdbDatabase, config.InfluxdbUser, config.InfluxdbPassword)
	influxdb.SetDebug(config.Debug)
	err := influxdb.Connect()
	check(err)
	metric := config.Metric

	filters := new(influxdbclient.Filters)

	filters.Add("host", config.StatsHost, "text")

	if len(config.StatsFilter) > 0 {
		filters.Add("name", config.StatsFilter, "regexp")
	}
	fromUnix, _ := ConvertTimeStamp(config.StatsFrom, config.Timezone)
	fromTime := fromUnix.Format(querytimeformat)
	toUnix, _ := ConvertTimeStamp(config.StatsTo, config.Timezone)
	toTime := toUnix.Format(querytimeformat)
	result, err := influxdb.ReadPoints("value", filters, "name", metric, fromTime, toTime, "")
	if err != nil {
		check(err)
	}

	//generate stats
	stats := influxdbclient.BuildStats(result)

	DisplayStats(&stats, config.StatsSort, config.StatsLimit)
}

func DisplayStats(stats *influxdbclient.DataStats, sort string, limit int) {
	fmt.Printf("%20s|%10s|%10s|%10s|%10s|%10s\n", "field", "Min", "Mean", "Median", "Max", "Points #")
	stats.FieldSort(sort)
	for i, stat := range *stats {
		fmt.Printf("%20s|%10.2f|%10.2f|%10.2f|%10.2f|%8d\n", stat.Name, stat.Min, stat.Mean, stat.Median, stat.Max, stat.Length)
		if limit > 0 {
			if i > limit {
				break
			}
		}
	}
}
