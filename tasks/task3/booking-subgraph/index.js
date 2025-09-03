import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { buildSubgraphSchema } from '@apollo/subgraph';
import gql from 'graphql-tag';

const typeDefs = gql`
    type Booking @key(fields: "id") {
        id: ID!
        userId: String!
        hotelId: String!
        promoCode: String
        discountPercent: Int
        hotel: Hotel  # Добавляем поле hotel
    }

    extend type Hotel @key(fields: "id") {
        id: ID! @external
    }

    type Query {
        bookingsByUser(userId: String!): [Booking]
    }
`;

const resolvers = {
    Query: {
        bookingsByUser: async (_, { userId }, { req }) => {
            // ACL проверка
            const requestingUserId = req.headers['userid'];
            if (!requestingUserId) {
                throw new Error('Unauthorized: userid header required');
            }

            // Проверяем, что пользователь запрашивает свои собственные бронирования
            if (requestingUserId !== userId) {
                throw new Error('Forbidden: You can only access your own bookings');
            }

            console.log(`ACL: User ${requestingUserId} accessing bookings for user ${userId}`);

            // Заглушка с тестовыми данными
            return [
                {
                    id: '1',
                    userId: userId,
                    hotelId: 'hotel-123',
                    promoCode: 'SUMMER25',
                    discountPercent: 15
                },
                {
                    id: '2',
                    userId: userId,
                    hotelId: 'hotel-456',
                    promoCode: null,
                    discountPercent: null
                }
            ];
        },
    },
    Booking: {
        __resolveReference: async (booking, { req }) => {
            // ACL проверка для reference resolution
            const requestingUserId = req.headers['userid'];
            if (!requestingUserId) {
                throw new Error('Unauthorized: userid header required for reference resolution');
            }

            console.log(`ACL: User ${requestingUserId} resolving booking reference: ${booking.id}`);

            // Заглушка для разрешения ссылок из других подграфов
            const resolvedBooking = {
                id: booking.id,
                userId: requestingUserId,
                hotelId: 'hotel-' + Math.floor(Math.random() * 1000),
                promoCode: Math.random() > 0.5 ? 'PROMO' + Math.floor(Math.random() * 100) : null,
                discountPercent: Math.random() > 0.5 ? Math.floor(Math.random() * 30) : null
            };

            // Дополнительная проверка: возвращаем бронирование только если оно принадлежит пользователю
            if (resolvedBooking.userId !== requestingUserId) {
                console.warn(`Forbidden: User ${requestingUserId} tried to access booking of user ${resolvedBooking.userId}`);
                return null;
            }

            return resolvedBooking;
        },

        // Добавляем резолвер для поля hotel
        hotel: (booking) => {
            // Возвращаем reference на отель, который будет разрешен hotel subgraph'ом
            return { __typename: 'Hotel', id: booking.hotelId };
        }
    },
};

const server = new ApolloServer({
    schema: buildSubgraphSchema([{ typeDefs, resolvers }]),
});

startStandaloneServer(server, {
    listen: { port: 4001 },
    context: async ({ req }) => ({ req }),
}).then(() => {
    console.log('✅ Booking subgraph ready at http://localhost:4001/');
});