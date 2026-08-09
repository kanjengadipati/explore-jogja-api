CREATE TABLE IF NOT EXISTS sales_commissions (
    id BIGSERIAL PRIMARY KEY,
    sales_user_id BIGINT NOT NULL,
    partner_user_id BIGINT NOT NULL,
    payment_transaction_id BIGINT NOT NULL,
    order_id VARCHAR(100) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,
    gross_amount DOUBLE PRECISION NOT NULL,
    commission_rate DOUBLE PRECISION NOT NULL,
    commission_amount DOUBLE PRECISION NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sales_commissions_sales_user ON sales_commissions(sales_user_id);
CREATE INDEX IF NOT EXISTS idx_sales_commissions_partner_user ON sales_commissions(partner_user_id);
CREATE INDEX IF NOT EXISTS idx_sales_commissions_status ON sales_commissions(status);
CREATE INDEX IF NOT EXISTS idx_sales_commissions_subject_type ON sales_commissions(subject_type);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_commissions_payment_tx ON sales_commissions(payment_transaction_id);
