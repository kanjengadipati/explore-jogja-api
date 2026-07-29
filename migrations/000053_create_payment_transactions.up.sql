CREATE TABLE IF NOT EXISTS payment_transactions (
    id SERIAL PRIMARY KEY,
    order_id TEXT NOT NULL UNIQUE,
    subject_type TEXT NOT NULL,
    subject_external_id TEXT NOT NULL,

    amount DOUBLE PRECISION NOT NULL,
    currency TEXT NOT NULL DEFAULT 'IDR',

    status TEXT NOT NULL DEFAULT 'pending',
    midtrans_token TEXT,
    payment_type TEXT,
    transaction_id TEXT,
    va_number TEXT,
    fraud_status TEXT,

    raw_notification TEXT,
    paid_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,

    created_by_user_id INTEGER,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_payment_transactions_subject ON payment_transactions (subject_type, subject_external_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions (status);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_deleted_at ON payment_transactions (deleted_at);
