package seeds

import (
	"log"
	"time"

	"pleco-api/internal/modules/article"

	"gorm.io/gorm"
)

func SeedArticles(db *gorm.DB) {
	mustHaveDB(db)

	now := time.Now()
	articles := []article.Article{
		{
			ExternalID:  "art-1",
			Slug:        "10-destinasi-wisata-wajib-jogja",
			Title:       "10 Destinasi Wisata Wajib Dikunjungi di Yogyakarta",
			TitleEn:     "10 Must-Visit Destinations in Yogyakarta",
			Excerpt:     "Dari Candi Prambanan hingga pantai selatan, temukan 10 destinasi terbaik yang tidak boleh kamu lewatkan saat berkunjung ke Yogyakarta.",
			ExcerptEn:   "From Prambanan Temple to the southern beaches, discover the 10 best destinations you should not miss when visiting Yogyakarta.",
			Content:     art1ContentID,
			ContentEn:   art1ContentEN,
			CoverImage:  "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b8/Prambanan_temple_compound%2C_2014-06-18.jpg/1280px-Prambanan_temple_compound%2C_2014-06-18.jpg",
			Category:    "panduan",
			Author:      "Jogjagem Team",
			Status:      "published",
			PublishedAt: &now,
			ReadTimeMinutes: 8,
			SeoTitle:    "10 Destinasi Wisata Wajib di Yogyakarta | Jogjagem",
			SeoTitleEn:  "10 Must-Visit Destinations in Yogyakarta | Jogjagem",
			SeoDescription:   "Panduan lengkap 10 destinasi wisata terbaik di Yogyakarta: Prambanan, Malioboro, Pantai Parangtritis, Gunung Merapi, dan lebih banyak lagi.",
			SeoDescriptionEn: "Complete guide to the 10 best tourist destinations in Yogyakarta: Prambanan, Malioboro, Parangtritis Beach, Mount Merapi, and more.",
			SeoKeywords:      "wisata Yogyakarta, destinasi jogja, Prambanan, Malioboro, Parangtritis, Gunung Merapi, panduan wisata jogja",
			SeoKeywordsEn:    "Yogyakarta tourism, jogja destinations, Prambanan, Malioboro, Parangtritis, Mount Merapi, Yogyakarta travel guide",
		},
		{
			ExternalID:  "art-2",
			Slug:        "itinerary-3-hari-jogja-budget",
			Title:       "Itinerary 3 Hari di Yogyakarta dengan Budget Hemat",
			TitleEn:     "3-Day Yogyakarta Itinerary on a Budget",
			Excerpt:     "Rencanakan liburan seru 3 hari di Jogja tanpa menguras kantong. Panduan lengkap dari transportasi, akomodasi, hingga kuliner.",
			ExcerptEn:   "Plan an exciting 3-day trip to Jogja without breaking the bank. Complete guide from transportation and accommodation to local food.",
			Content:     art2ContentID,
			ContentEn:   art2ContentEN,
			CoverImage:  "https://upload.wikimedia.org/wikipedia/commons/thumb/2/22/Borobudur-Nothwest-Bird%27s-eye-view.jpg/1280px-Borobudur-Nothwest-Bird%27s-eye-view.jpg",
			Category:    "itinerary",
			Author:      "Jogjagem Team",
			Status:      "published",
			PublishedAt: &now,
			ReadTimeMinutes: 10,
			SeoTitle:    "Itinerary 3 Hari Yogyakarta Budget Hemat | Jogjagem",
			SeoTitleEn:  "3-Day Budget Yogyakarta Itinerary | Jogjagem",
			SeoDescription:   "Itinerary lengkap 3 hari di Yogyakarta dengan budget hemat: hotel murah, transportasi, wisata gratis dan berbayar, serta kuliner wajib coba.",
			SeoDescriptionEn: "Complete 3-day budget itinerary for Yogyakarta: affordable hotels, transport tips, free and paid attractions, and must-try local food.",
			SeoKeywords:      "itinerary jogja 3 hari, wisata jogja murah, budget jogja, backpacker jogja, paket wisata jogja hemat",
			SeoKeywordsEn:    "Yogyakarta 3 day itinerary, budget Yogyakarta trip, cheap Jogja travel, backpacker Jogja, affordable Yogyakarta tour",
		},
		{
			ExternalID:  "art-3",
			Slug:        "kuliner-legendaris-jogja",
			Title:       "10 Kuliner Legendaris Yogyakarta yang Wajib Dicoba",
			TitleEn:     "10 Legendary Yogyakarta Foods You Must Try",
			Excerpt:     "Gudeg, bakpia, sate klathak, dan berbagai kuliner khas Yogyakarta yang sudah ada sejak puluhan tahun. Di mana bisa menemukannya?",
			ExcerptEn:   "Gudeg, bakpia, sate klathak, and other iconic Yogyakarta dishes that have been around for decades. Where can you find them?",
			Content:     art3ContentID,
			ContentEn:   art3ContentEN,
			CoverImage:  "https://upload.wikimedia.org/wikipedia/commons/thumb/1/16/Gudeg_Jogja.jpg/1280px-Gudeg_Jogja.jpg",
			Category:    "kuliner",
			Author:      "Jogjagem Team",
			Status:      "published",
			PublishedAt: &now,
			ReadTimeMinutes: 7,
			SeoTitle:    "10 Kuliner Legendaris Yogyakarta Wajib Coba | Jogjagem",
			SeoTitleEn:  "10 Legendary Yogyakarta Foods to Try | Jogjagem",
			SeoDescription:   "Daftar kuliner legendaris Yogyakarta yang wajib dicoba: gudeg Bu Tjitro, bakpia Pathuk, sate klathak Jejeran, angkringan, dan masih banyak lagi.",
			SeoDescriptionEn: "List of legendary Yogyakarta foods you must try: gudeg, bakpia Pathuk, sate klathak, angkringan street food, and many more.",
			SeoKeywords:      "kuliner jogja, makanan khas yogyakarta, gudeg jogja, bakpia pathuk, sate klathak, wisata kuliner jogja",
			SeoKeywordsEn:    "Yogyakarta food, traditional Jogja cuisine, gudeg Yogyakarta, bakpia Pathuk, sate klathak, Yogyakarta food tour",
		},
		{
			ExternalID:  "art-4",
			Slug:        "hidden-gems-yogyakarta",
			Title:       "Hidden Gems Yogyakarta: Destinasi Tersembunyi yang Belum Ramai",
			TitleEn:     "Hidden Gems of Yogyakarta: Off-the-Beaten-Path Destinations",
			Excerpt:     "Bosan dengan destinasi mainstream? Temukan tempat-tempat tersembunyi di Yogyakarta yang masih sepi pengunjung namun menyimpan keindahan luar biasa.",
			ExcerptEn:   "Tired of mainstream spots? Discover Yogyakarta's hidden gems that are still crowd-free yet hold extraordinary beauty.",
			Content:     art4ContentID,
			ContentEn:   art4ContentEN,
			CoverImage:  "https://upload.wikimedia.org/wikipedia/commons/thumb/a/a7/Yogyakarta_kraton_2.jpg/1280px-Yogyakarta_kraton_2.jpg",
			Category:    "panduan",
			Author:      "Jogjagem Team",
			Status:      "published",
			PublishedAt: &now,
			ReadTimeMinutes: 9,
			SeoTitle:    "Hidden Gems Yogyakarta: Destinasi Tersembunyi Wajib Dikunjungi | Jogjagem",
			SeoTitleEn:  "Hidden Gems of Yogyakarta: Secret Destinations to Visit | Jogjagem",
			SeoDescription:   "Temukan hidden gems Yogyakarta yang belum banyak diketahui wisatawan: bukit tersembunyi, gua alami, air terjun terpencil, dan desa wisata unik.",
			SeoDescriptionEn: "Discover Yogyakarta's hidden gems unknown to most tourists: secret hills, natural caves, secluded waterfalls, and unique cultural villages.",
			SeoKeywords:      "hidden gems jogja, destinasi tersembunyi yogyakarta, wisata sepi jogja, tempat wisata unik jogja, off the beaten path jogja",
			SeoKeywordsEn:    "hidden gems Yogyakarta, secret places Jogja, off the beaten path Yogyakarta, undiscovered Jogja destinations",
		},
		{
			ExternalID:  "art-5",
			Slug:        "tips-wisata-prambanan",
			Title:       "Tips Wisata ke Candi Prambanan: Panduan Lengkap 2025",
			TitleEn:     "Prambanan Temple Travel Tips: Complete Guide 2025",
			Excerpt:     "Segala yang perlu kamu tahu sebelum mengunjungi Candi Prambanan: harga tiket, jam terbaik, cara ke sana, dan hal-hal yang sering terlewat wisatawan.",
			ExcerptEn:   "Everything you need to know before visiting Prambanan Temple: ticket prices, best visiting hours, how to get there, and things most tourists miss.",
			Content:     art5ContentID,
			ContentEn:   art5ContentEN,
			CoverImage:  "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b8/Prambanan_temple_compound%2C_2014-06-18.jpg/1280px-Prambanan_temple_compound%2C_2014-06-18.jpg",
			Category:    "panduan",
			Author:      "Jogjagem Team",
			Status:      "published",
			PublishedAt: &now,
			ReadTimeMinutes: 6,
			SeoTitle:    "Tips Wisata Candi Prambanan 2025: Tiket, Jam Buka & Cara ke Sana | Jogjagem",
			SeoTitleEn:  "Prambanan Temple Guide 2025: Tickets, Opening Hours & How to Get There | Jogjagem",
			SeoDescription:   "Panduan lengkap wisata Candi Prambanan 2025: harga tiket masuk, jam buka, cara ke sana dari Jogja, tips foto terbaik, dan apa yang harus dibawa.",
			SeoDescriptionEn: "Complete guide to Prambanan Temple 2025: entrance fees, opening hours, how to get there from Jogja, best photo spots, and what to bring.",
			SeoKeywords:      "Candi Prambanan, tiket prambanan, jam buka prambanan, cara ke prambanan dari jogja, wisata prambanan 2025",
			SeoKeywordsEn:    "Prambanan Temple, Prambanan ticket price, Prambanan opening hours, how to get to Prambanan from Jogja, Prambanan visit guide 2025",
		},
		{
			ExternalID:  "art-6",
			Slug:        "wisata-malioboro-panduan-lengkap",
			Title:       "Panduan Lengkap Wisata di Malioboro Yogyakarta",
			TitleEn:     "Complete Guide to Malioboro Yogyakarta",
			Excerpt:     "Malioboro bukan sekadar jalan belanja. Pelajari sejarah, spot foto terbaik, tempat makan lokal, dan cara menikmati Malioboro seperti warga lokal.",
			ExcerptEn:   "Malioboro is more than just a shopping street. Learn its history, best photo spots, local eateries, and how to experience it like a local.",
			Content:     art6ContentID,
			ContentEn:   art6ContentEN,
			CoverImage:  "https://upload.wikimedia.org/wikipedia/commons/thumb/7/7e/Malioboro_Street_Yogyakarta.jpg/1280px-Malioboro_Street_Yogyakarta.jpg",
			Category:    "panduan",
			Author:      "Jogjagem Team",
			Status:      "published",
			PublishedAt: &now,
			ReadTimeMinutes: 7,
			SeoTitle:    "Panduan Wisata Malioboro Yogyakarta: Tips, Belanja & Kuliner | Jogjagem",
			SeoTitleEn:  "Malioboro Yogyakarta Travel Guide: Tips, Shopping & Food | Jogjagem",
			SeoDescription:   "Panduan wisata Malioboro Yogyakarta: belanja batik dan souvenir, kuliner khas, angkringan, hotel terdekat, dan tips agar tidak tertipu pedagang.",
			SeoDescriptionEn: "Malioboro Yogyakarta travel guide: batik and souvenir shopping, local street food, angkringan, nearby hotels, and tips to avoid tourist traps.",
			SeoKeywords:      "Malioboro Yogyakarta, wisata Malioboro, belanja Malioboro, kuliner Malioboro, panduan Malioboro, jalan Malioboro",
			SeoKeywordsEn:    "Malioboro Yogyakarta, visit Malioboro, shopping Malioboro, Malioboro food, Malioboro travel guide, Malioboro street",
		},
	}

	for _, a := range articles {
		existing := &article.Article{}
		err := db.Where("external_id = ?", a.ExternalID).First(existing).Error
		if err == nil {
			// Already exists — skip to preserve any edits made via admin
			continue
		}
		if err := db.Create(&a).Error; err != nil {
			log.Printf("⚠️  seed article %s: %v", a.ExternalID, err)
		}
	}
	log.Printf("✅ seeded %d articles", len(articles))
}
