package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/urfave/cli/v2"

	"github.com/atscan/blobd/blob"
)

func main() {

	app := &cli.App{
		Name:  "blobd",
		Usage: "AT Protocol Blob-serving HTTP Server in Go",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   3000,
			},
			&cli.StringFlag{
				Name:  "host",
				Value: "127.0.0.1",
			},
			&cli.StringFlag{
				Name:    "data-dir",
				Aliases: []string{"d"},
				Value:   "./data",
			},
		},
		Action: serve,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func serve(cctx *cli.Context) error {
	server := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	server.Use(cors.New())

	dir := cctx.String("data-dir")
	fmt.Printf("Data directory: %v\n", dir)

	server.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("blobd active")
	})
	server.Get("/:did/:cid", func(c *fiber.Ctx) error {
		return response(c, dir, nil, nil)
	})
	server.Get("/:did/:cid/inspect", func(c *fiber.Ctx) error {
		return inspect(c, dir)
	})
	listen := cctx.String("host") + ":" + cctx.String("port")
	fmt.Printf("blobd started at %v\n", listen)
	log.Fatal(server.Listen(listen))
	return nil
}

func getBlobIndex(c *fiber.Ctx, dir string) (blob.Blob, error) {
	did, cid := c.Params("did"), c.Params("cid")
	if strings.Index(cid, "baf") != 0 {
		return blob.Blob{}, errors.New("Bad cid")
	}
	blob, err := blob.Get(dir, did, cid)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return blob, errors.New("Cannot get blob")
	}
	return blob, nil
}

func response(c *fiber.Ctx, dir string, of *string, ofc *blob.OutputFormatOptions) error {
	start := time.Now()
	b, err := getBlobIndex(c, dir)
	if err != nil {
		return c.SendStatus(404)
	}
	m := c.Queries()
	if of == nil {
		fmt := m["format"]
		if fmt != "" {
			of = &fmt
		} else {
			fmt = "raw"
			of = &fmt
		}
	}
	if ofc == nil {
		ofc = &blob.OutputFormatOptions{}
		if m["width"] != "" {
			ofc.Width, _ = strconv.Atoi(m["width"])
		}
	}
	out, err := b.Output(dir, of, ofc)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return c.SendStatus(500)
	}
	c.Set("Content-Type", out.ContentType)

	fmt.Printf("%v %v [%v]\n", c.Method(), c.Path(), time.Since(start))
	return c.Send(out.Body())
}

func inspect(c *fiber.Ctx, dir string) error {
	b, err := getBlobIndex(c, dir)
	if err != nil {
		return c.SendStatus(404)
	}
	js, err := json.MarshalIndent(b, "", "  ")
	fmt.Printf("%v %v\n", c.Method(), c.Path())
	return c.Send([]byte(js))
}
