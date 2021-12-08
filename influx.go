package smarthome

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxClient struct {
	c        client.HTTPClient
	database string
}

//		Addr:     "https://example.com:8086",
//		Username: "admin",
//		Password: "secret",

func NewInfluxClient(addr string, user string, pass string, database string) (*InfluxClient, error) {
	var influxClient InfluxClient
	var err error

	influxClient.c, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: user,
		Password: pass,
	})
	if err != nil {
		return nil, err
	}
	influxClient.database = database
	return &influxClient, nil
}

func (i *InfluxClient) Insert(nameTags string, fields map[string]interface{}) error {
	names := strings.Split(nameTags, ",")
	tags := make(map[string]string)
	for _, t := range names[1:] {
		res := strings.Split(t, "=")
		if len(res) != 2 {
			return fmt.Errorf("InfluxClient.Insert(): syntax error in nameTags=%q", nameTags)
		}
		tags[res[0]] = res[1]
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	pt, err := client.NewPoint(names[0], tags, fields, time.Now())
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	return i.c.Write(bp)
}

func (i *InfluxClient) InsertLine(line string) error {
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		return fmt.Errorf("InfluxClient.InsertLine: syntax error")
	}
	names := strings.Split(parts[0], ",")
	tags := make(map[string]string)
	for _, t := range names[1:] {
		res := strings.Split(t, "=")
		if len(res) != 2 {
			return fmt.Errorf("InfluxClient.InsertLine: syntax error in tags: %q", parts[0])
		}
		tags[res[0]] = res[1]
	}

	fields := make(map[string]interface{})
	for _, f := range strings.Split(parts[1], ",") {
		res := strings.Split(f, "=")
		if len(res) != 2 {
			return fmt.Errorf("Influxclient.InsertLine: syntax error in fields: %q", parts[1])
		}
		i, err := strconv.Atoi(res[1])
		if err == nil {
			fields[res[0]] = i
			continue
		}
		f, err := strconv.ParseFloat(res[1], 64)
		if err == nil {
			fields[res[0]] = f
			continue
		}
		fields[res[0]] = res[1]
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	pt, err := client.NewPoint(names[0], tags, fields, time.Now())
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	return i.c.Write(bp)
}

func (i *InfluxClient) Close() error {
	return i.c.Close()
}
