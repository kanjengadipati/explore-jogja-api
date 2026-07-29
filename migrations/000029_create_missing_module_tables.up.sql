-- Hotels
CREATE TABLE hotels (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255),
    address VARCHAR(500),
    stars INTEGER DEFAULT 0,
    price_per_night VARCHAR(100),
    images JSONB DEFAULT '[]',
    amenities JSONB DEFAULT '[]',
    phone VARCHAR(50),
    email VARCHAR(255),
    website VARCHAR(500),
    rating DECIMAL(3,1) DEFAULT 0,
    review_count INTEGER DEFAULT 0,
    latitude DECIMAL(10,6),
    longitude DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_hotels_external_id ON hotels(external_id);
CREATE INDEX idx_hotels_deleted_at ON hotels(deleted_at);

-- Restaurants
CREATE TABLE restaurants (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255),
    address VARCHAR(500),
    cuisine_type VARCHAR(100),
    price_range VARCHAR(100),
    images JSONB DEFAULT '[]',
    opening_hours VARCHAR(255),
    phone VARCHAR(50),
    rating DECIMAL(3,1) DEFAULT 0,
    review_count INTEGER DEFAULT 0,
    latitude DECIMAL(10,6),
    longitude DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_restaurants_external_id ON restaurants(external_id);
CREATE INDEX idx_restaurants_deleted_at ON restaurants(deleted_at);

-- Guides
CREATE TABLE guides (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    bio TEXT,
    specialization VARCHAR(255),
    phone VARCHAR(50),
    email VARCHAR(255),
    rating DECIMAL(3,1) DEFAULT 0,
    review_count INTEGER DEFAULT 0,
    languages JSONB DEFAULT '[]',
    price_per_day VARCHAR(100),
    avatar VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_guides_external_id ON guides(external_id);
CREATE INDEX idx_guides_deleted_at ON guides(deleted_at);

-- Souvenirs
CREATE TABLE souvenirs (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255),
    address VARCHAR(500),
    images JSONB DEFAULT '[]',
    product_types JSONB DEFAULT '[]',
    price_range VARCHAR(100),
    phone VARCHAR(50),
    rating DECIMAL(3,1) DEFAULT 0,
    latitude DECIMAL(10,6),
    longitude DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_souvenirs_external_id ON souvenirs(external_id);
CREATE INDEX idx_souvenirs_deleted_at ON souvenirs(deleted_at);

-- Rentals
CREATE TABLE rentals (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255),
    address VARCHAR(500),
    vehicle_types JSONB DEFAULT '[]',
    price_per_day VARCHAR(100),
    images JSONB DEFAULT '[]',
    phone VARCHAR(50),
    rating DECIMAL(3,1) DEFAULT 0,
    latitude DECIMAL(10,6),
    longitude DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_rentals_external_id ON rentals(external_id);
CREATE INDEX idx_rentals_deleted_at ON rentals(deleted_at);

-- Stories
CREATE TABLE stories (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    user_id VARCHAR(100),
    title VARCHAR(500) NOT NULL,
    content TEXT,
    images JSONB DEFAULT '[]',
    destination_ids JSONB DEFAULT '[]',
    likes INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_stories_external_id ON stories(external_id);
CREATE INDEX idx_stories_user_id ON stories(user_id);
CREATE INDEX idx_stories_status ON stories(status);
CREATE INDEX idx_stories_deleted_at ON stories(deleted_at);

-- Promotions
CREATE TABLE promotions (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    discount VARCHAR(50),
    start_date VARCHAR(50),
    end_date VARCHAR(50),
    image_url VARCHAR(500),
    category VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    code VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE INDEX idx_promotions_external_id ON promotions(external_id);
CREATE INDEX idx_promotions_category ON promotions(category);
CREATE INDEX idx_promotions_status ON promotions(status);
CREATE INDEX idx_promotions_deleted_at ON promotions(deleted_at);
