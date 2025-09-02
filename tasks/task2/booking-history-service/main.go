package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

type BookingResponse struct {
	Id              string  `json:"Id"`
	UserId          string  `json:"UserId"`
	HotelId         string  `json:"HotelId"`
	PromoCode       string  `json:"PromoCode"`
	DiscountPercent float64 `json:"DiscountPercent"`
	Price           float64 `json:"Price"`
	CreatedAt       string  `json:"CreatedAt"` // ISO-8601
}

func main() {
	// Фиксированные параметры
	kafkaURL := "kafka:9092"
	topic := "bookings"
	groupID := "booking-history-group"

	fmt.Printf("Starting booking-history-service...\n")
	fmt.Printf("Kafka: %s, Topic: %s, Group: %s\n", kafkaURL, topic, groupID)

	// Создание Kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{kafkaURL},
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
	})
	defer reader.Close()

	// Обработка сигналов для graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Waiting for messages... Press Ctrl+C to exit.")

	// Бесконечный цикл чтения сообщений
	for {
		select {
		case <-sigchan:
			fmt.Println("\nShutting down...")
			return
		default:
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message: %v", err)
				time.Sleep(1 * time.Second) // Пауза при ошибках
				continue
			}

			// Десериализация JSON
			var booking BookingResponse
			if err := json.Unmarshal(msg.Value, &booking); err != nil {
				log.Printf("Error parsing JSON: %v", err)
				log.Printf("Raw message: %s", string(msg.Value))
				continue
			}

			// Вывод информации о бронировании
			fmt.Printf("\n=== New Booking Received ===\n")
			fmt.Printf("Offset: %d, Partition: %d\n", msg.Offset, msg.Partition)
			fmt.Printf("Booking ID: %s\n", booking.Id)
			fmt.Printf("User ID: %s\n", booking.UserId)
			fmt.Printf("Hotel ID: %s\n", booking.HotelId)
			fmt.Printf("Promo Code: %s\n", booking.PromoCode)
			fmt.Printf("Discount: %.2f%%\n", booking.DiscountPercent)
			fmt.Printf("Price: $%.2f\n", booking.Price)
			fmt.Printf("Created At: %s\n", booking.CreatedAt)
			fmt.Printf("=============================\n")
		}
	}
}
