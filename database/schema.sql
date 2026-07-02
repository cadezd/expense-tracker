CREATE TABLE IF NOT EXISTS users (
    -- FIELD DEFINITIONS
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    object_version BIGINT NOT NULL DEFAULT 1 
        CHECK(object_version >= 1),

    -- CONSTRAINTS
    CONSTRAINT uq_email
        UNIQUE (email)
);


CREATE TABLE IF NOT EXISTS receipts (
    -- FIELD DEFINITIONS
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    original_filename TEXT NOT NULL,
    stored_filename TEXT NOT NULL,
    storage_path TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size BIGINT 
        CHECK(file_size >= 0),
    status TEXT NOT NULL DEFAULT 'uploaded'
        CHECK (
            status IN (
                'uploaded',
                'processing',
                'processed',
                'failed'
            )
        ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    object_version BIGINT NOT NULL DEFAULT 1
        CHECK(object_version >= 1), 

    -- CONSTRAINTS
    CONSTRAINT fk_receipts_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_receipts_user_id_id
        UNIQUE (user_id, id)
);

CREATE TABLE IF NOT EXISTS transactions (
    -- FIELD DEFINITIONS
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    receipt_id UUID,
    type TEXT NOT NULL
        CHECK (type IN ('expense', 'income')),
    counterparty TEXT,
    amount NUMERIC(12, 2) NOT NULL 
        CHECK (amount > 0),
    currency CHAR(3) NOT NULL DEFAULT 'EUR',
    transaction_date DATE NOT NULL DEFAULT CURRENT_DATE,
    category TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    object_version BIGINT NOT NULL DEFAULT 1
        CHECK(object_version >= 1), 

    -- CONSTRAINTS
    CONSTRAINT fk_transactions_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_transactions_receipt_owner
        FOREIGN KEY (user_id, receipt_id)
        REFERENCES receipts(user_id, id)
        ON DELETE SET NULL (receipt_id),

    CONSTRAINT uq_transactions_receipt
        UNIQUE (receipt_id)
);