CREATE TYPE order_status AS ENUM ('created', 'validated', 'validation failed', 'payment pending', 'paid', 'payment failed', 'cancelled', 'confirmed');

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    amount INT NOT NULL,
    status order_status NOT NULL DEFAULT 'created',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


