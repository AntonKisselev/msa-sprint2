package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/kafka-go"

	pb "booking/booking"

	"google.golang.org/grpc"
)

var conn *pgx.Conn
var writer *kafka.Writer

type server struct {
	pb.UnimplementedBookingServiceServer
}

func (s *server) CreateBooking(ctx context.Context, req *pb.BookingRequest) (*pb.BookingResponse, error) {
	log.Printf("Received: CreateBooking %v", req)

	created := time.Now()
	booking := &pb.BookingResponse{
		Id:              "",
		UserId:          req.UserId,
		HotelId:         req.HotelId,
		PromoCode:       req.PromoCode,
		DiscountPercent: 0,
		Price:           0,
		CreatedAt:       created.Format(time.RFC3339),
	}

	err := conn.QueryRow(ctx, "INSERT INTO booking (user_id, hotel_id, promo_code, discount_percent, price, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		booking.UserId, booking.HotelId, booking.PromoCode, booking.DiscountPercent, booking.Price, booking.CreatedAt).Scan(&booking.Id)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Booking added id = ", booking.Id)
	err = writeKafka(booking)
	if err != nil {
		log.Fatal(err)
	}

	return booking, nil
}

func (s *server) ListBookings(ctx context.Context, req *pb.BookingListRequest) (*pb.BookingListResponse, error) {
	log.Printf("Received: ListBookings %v", req)

	bq := squirrel.Select("id", "user_id", "hotel_id", "promo_code", "discount_percent", "price", "created_at").
		From("booking")
	if req.UserId != "" {
		bq = bq.Where(squirrel.Eq{"user_id": req.UserId})
	}

	query, args, err := bq.OrderBy("created_at DESC").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("SQL:", query)
	log.Println("Args:", args)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var bookings []*pb.BookingResponse
	for rows.Next() {
		var booking pb.BookingResponse
		var createdAt time.Time
		if err = rows.Scan(&booking.Id, &booking.UserId, &booking.HotelId, &booking.PromoCode, &booking.DiscountPercent, &booking.Price, &createdAt); err != nil {
			return nil, err
		}
		booking.CreatedAt = createdAt.Format(time.RFC3339)
		bookings = append(bookings, &booking)
	}

	res := &pb.BookingListResponse{
		Bookings: bookings,
	}
	return res, nil
}

func writeKafka(booking *pb.BookingResponse) error {
	// Сериализация в JSON
	messageBytes, err := json.Marshal(booking)
	if err != nil {
		return err
	}

	// Создание Kafka сообщения
	msg := kafka.Message{
		Key:   []byte(booking.Id),
		Value: messageBytes,
	}

	// Отправка сообщения
	err = writer.WriteMessages(context.Background(), msg)
	if err != nil {
		return err
	}

	log.Println("Message sent successfully to kafka")
	return nil
}

func main() {
	// Подключение к базе данных
	connStr := "postgres://hotelio-booking:hotelio-booking@booking-db:5432/hotelio-booking"
	var err error
	conn, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	/// kafka connect
	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{"kafka:9092"},
		Topic:    "bookings",
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	s := grpc.NewServer()
	pb.RegisterBookingServiceServer(s, &server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
