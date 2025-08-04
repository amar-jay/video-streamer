// Package main contains an example.
package main

import (
	"fmt"
	"log"
	"matek-video-streamer/internal/server"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "nebula-video-streamer",
		Usage: "RTSP video streamer",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Value:    "/home/amarjay/Desktop/code/video-streamer/camera_stream.h264",
				Usage:    "Path to the input video file",
				Required: false,
			},
			&cli.StringFlag{
				Name:  "rtsp-address",
				Value: ":8554",
				Usage: "RTSP server address",
			},
			&cli.StringFlag{
				Name:  "udp-rtp-address",
				Value: ":8000",
				Usage: "UDP RTP address",
			},
			&cli.StringFlag{
				Name:  "udp-rtcp-address",
				Value: ":8001",
				Usage: "UDP RTCP address",
			},
		},
		Action: func(c *cli.Context) error {
			inputFile := c.String("input")

			// Check if the input file exists
			if _, err := os.Stat(inputFile); os.IsNotExist(err) {
				return fmt.Errorf("input file does not exist: %s", inputFile)
			}

			log.Printf("Starting video streamer with input: %s", inputFile)
			return server.StartServer(inputFile, c.String("rtsp-address"), c.String("udp-rtp-address"), c.String("udp-rtcp-address"))
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
