package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/colinmarc/hdfs"
	"github.com/mdouchement/iosupport"
)

func open(path string) (iosupport.FileReader, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("TsvSorter input: %v", err)
	}

	switch u.Scheme {
	case "hdfs":
		client, err := hdfs.New(u.Host)
		if err != nil {
			return nil, fmt.Errorf("TsvSorter input (hdfs): %v", err)
		}
		return client.Open(u.Path)
	case "":
		return os.Open(u.Path)
	default:
		return nil, fmt.Errorf("TsvSorter: Usupporting input scheme: %s", u.Scheme)
	}
}

func create(path string) (iosupport.FileWriter, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("TsvSorter output: %v", err)
	}

	switch u.Scheme {
	case "hdfs":
		client, err := hdfs.New(u.Host)
		if err != nil {
			return nil, fmt.Errorf("TsvSorter output (hdfs): %v", err)
		}
		return client.Create(u.Path)
	case "":
		return os.Create(u.Path)
	default:
		return nil, fmt.Errorf("TsvSorter: Usupporting output scheme: %s", u.Scheme)
	}
}
