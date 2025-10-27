CREATE TABLE IF NOT EXISTS analyst_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticker STRING NOT NULL,
    company STRING NOT NULL,
    brokerage STRING,
    action STRING NOT NULL,
    rating_from STRING,
    rating_to STRING,
    target_from DECIMAL(10, 2),
    target_to DECIMAL(10, 2),
    recommendation_date TIMESTAMP DEFAULT current_timestamp(),
    status BOOLEAN,
    created_at TIMESTAMP DEFAULT current_timestamp(),
    updated_at TIMESTAMP DEFAULT current_timestamp()
);

-- Index for faster queries by ticker
CREATE INDEX IF NOT EXISTS idx_ticker ON analyst_recommendations(ticker);

-- Index for filtering by action type
CREATE INDEX IF NOT EXISTS idx_action ON analyst_recommendations(action);

-- Index for date-based queries
CREATE INDEX IF NOT EXISTS idx_recommendation_date ON analyst_recommendations(recommendation_date DESC);