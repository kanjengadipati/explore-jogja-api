CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    title_en VARCHAR(500),
    excerpt TEXT,
    excerpt_en TEXT,
    content TEXT,
    content_en TEXT,
    cover_image VARCHAR(500),
    category VARCHAR(100),
    tags JSONB DEFAULT '[]',
    author VARCHAR(255),
    status VARCHAR(50) DEFAULT 'draft',
    published_at TIMESTAMP,
    seo_title VARCHAR(500),
    seo_title_en VARCHAR(500),
    seo_description TEXT,
    seo_description_en TEXT,
    seo_keywords TEXT,
    seo_keywords_en TEXT,
    og_image VARCHAR(500),
    read_time_minutes INTEGER DEFAULT 5,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_articles_external_id ON articles(external_id);
CREATE INDEX idx_articles_slug ON articles(slug);
CREATE INDEX idx_articles_category ON articles(category);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_published_at ON articles(published_at);
CREATE INDEX idx_articles_deleted_at ON articles(deleted_at);
