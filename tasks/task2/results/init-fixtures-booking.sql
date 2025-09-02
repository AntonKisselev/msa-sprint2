create table  IF NOT EXISTS booking
(
    discount_percent double precision,
    price            double precision not null,
    created_at       timestamp(6) with time zone,
                                      id               bigserial
                                      primary key,
                                      hotel_id         varchar(255),
    promo_code       varchar(255),
    user_id          varchar(255)
    );


DELETE FROM booking;
  
-- Бронирования (для GET /api/bookings)
INSERT INTO booking (user_id, hotel_id, promo_code, discount_percent, price, created_at)
VALUES
('test-user-2', 'test-hotel-1', 'TESTCODE1', 10.0, 90.0, NOW()),
('test-user-3', 'test-hotel-1', null, 0.0, 80.0, NOW());

