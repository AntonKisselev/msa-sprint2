import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { buildSubgraphSchema } from '@apollo/subgraph';
import gql from 'graphql-tag';

const typeDefs = gql`
    type Hotel @key(fields: "id") {
        id: ID!
        name: String
        city: String
        stars: Int
    }

    type Query {
        hotelsByIds(ids: [ID!]!): [Hotel]
    }
`;

const resolvers = {
    Hotel: {
        __resolveReference: async ({ id }, { req }) => {
            // ACL проверка
            const userId = req.headers['userid'];
            if (!userId) {
                throw new Error('Unauthorized: userid header required');
            }

            console.log(`ACL: User ${userId} resolving hotel reference: ${id}`);

            // Заглушка с тестовыми данными
            const hotels = {
                'hotel-123': { id: 'hotel-123', name: 'Grand Plaza Hotel', city: 'New York', stars: 5 },
                'hotel-456': { id: 'hotel-456', name: 'Seaside Resort', city: 'Miami', stars: 4 },
                'hotel-789': { id: 'hotel-789', name: 'Mountain Lodge', city: 'Denver', stars: 3 },
            };

            return hotels[id] || null;
        },
    },
    Query: {
        hotelsByIds: async (_, { ids }, { req }) => {
            // ACL проверка
            const userId = req.headers['userid'];
            if (!userId) {
                throw new Error('Unauthorized: userid header required');
            }

            console.log(`ACL: User ${userId} fetching hotels by IDs: ${ids.join(', ')}`);

            const hotelsDatabase = [
                { id: 'hotel-123', name: 'Grand Plaza Hotel', city: 'New York', stars: 5 },
                { id: 'hotel-456', name: 'Seaside Resort', city: 'Miami', stars: 4 },
                { id: 'hotel-789', name: 'Mountain Lodge', city: 'Denver', stars: 3 },
                { id: 'hotel-101', name: 'City Center Inn', city: 'Chicago', stars: 2 },
                { id: 'hotel-202', name: 'Airport Hotel', city: 'Los Angeles', stars: 3 },
            ];

            return hotelsDatabase.filter(hotel => ids.includes(hotel.id));
        },
    },
};

const server = new ApolloServer({
    schema: buildSubgraphSchema([{ typeDefs, resolvers }]),
});

startStandaloneServer(server, {
    listen: { port: 4002 },
    context: async ({ req }) => ({ req }),
}).then(() => {
    console.log('✅ Hotel subgraph ready at http://localhost:4002/');
});