package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/tuvisolutions/tuvi-website-backend/internal/apihttp"
	"github.com/tuvisolutions/tuvi-website-backend/internal/calendar"
	"github.com/tuvisolutions/tuvi-website-backend/internal/config"
	"github.com/tuvisolutions/tuvi-website-backend/internal/consultations"
	"github.com/tuvisolutions/tuvi-website-backend/internal/email"
	"github.com/tuvisolutions/tuvi-website-backend/internal/store"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer st.Close()

	var cal calendar.Provider
	if cfg.GoogleCalendarDisabled {
		log.Println("Google Calendar disabled — using DB-only mode")
		cal = calendar.NewNoop()
	} else {
		gcal, activeID, err := calendar.EnsureWritable(ctx, cfg.GoogleCalendarID, cfg.GoogleServiceAccountJSON)
		if err != nil {
			log.Fatalf("google calendar: %v", err)
		}
		if activeID != cfg.GoogleCalendarID {
			log.Printf("WARNING: configured calendar not writable — using %s", activeID)
			log.Printf("→ To use your team calendar, share %s with tuvi-solutions@iowe-f76af.iam.gserviceaccount.com (Make changes to events)", cfg.GoogleCalendarID)
		}
		if _, err := gcal.FreeBusy(ctx, time.Now(), time.Now().Add(time.Hour)); err != nil {
			log.Printf("WARNING: Google Calendar freebusy check: %v", err)
		} else {
			log.Println("Google Calendar connected:", activeID)
		}
		cal = gcal
	}

	mailer := email.NewSender(email.Config{
		FromAddress: cfg.EmailFromAddress,
		FromName:    cfg.EmailFromName,
		Host:        cfg.SMTPHost,
		Port:        cfg.SMTPPort,
		Username:    cfg.SMTPUsername,
		Password:    cfg.SMTPPassword,
		UseTLS:      cfg.SMTPUseTLS,
		Disabled:    cfg.EmailDisabled,
	})

	svc := consultations.NewService(cfg, st, cal, mailer)
	handlers := apihttp.NewHandlers(svc, cfg.APIToken)

	app := fiber.New(fiber.Config{
		AppName: "Tuvi Consultation API",
	})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://127.0.0.1:3000,http://localhost:3001,http://127.0.0.1:3001",
		AllowHeaders: "Authorization, Content-Type",
	}))

	handlers.Register(app)

	go func() {
		addr := cfg.HTTPAddr
		log.Printf("consultation API listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("shutting down...")
	_ = app.Shutdown()
}
