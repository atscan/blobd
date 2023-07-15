package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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
	dir := cctx.String("data-dir")
	fmt.Printf("Data directory: %v\n", dir)

	server.Get("/:did/:cid", func(c *fiber.Ctx) error {
		start := time.Now()
		did, cid := c.Params("did"), c.Params("cid")

		if strings.Index(cid, "baf") != 0 {
			return c.SendStatus(404)
		}
		r, _ := regexp.Compile("^([^\\.]+)(\\.([0-9]+)(x[0-9]+|)px|)\\.(webp)$")
		var of string = "raw"
		ofc := blob.OutputFormatOptions{}
		if m := r.FindStringSubmatch(cid); len(m) == 6 {
			cid = m[1]
			of = m[5]
			if m[3] != "" {
				ofc.Width, _ = strconv.Atoi(m[3])
			}
			/*if m[4] != "" {
				ofc.Height, _ = strconv.Atoi(m[4])
			}*/
		}

		// get blob
		blob, err := blob.Get(dir, did, cid)
		if err != nil {
			log.Println("Error: ", err)
			return c.SendStatus(404)
		}
		out, err := blob.Output(dir, of, ofc)
		if err != nil {
			return c.SendStatus(500)
		}

		c.Set("Content-Type", out.ContentType)

		fmt.Printf("%v %v [%v]\n", c.Method(), c.Path(), time.Since(start))
		return c.Send(out.Body())
	})

	server.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("blobd active")
	})

	listen := cctx.String("host") + ":" + cctx.String("port")
	fmt.Printf("blobd started at %v\n", listen)
	log.Fatal(server.Listen(listen))
	return nil
}
