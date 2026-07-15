package seeds

import (
	"fmt"
	"log"

	"pleco-api/internal/modules/partner"

	"gorm.io/gorm"
)

func SeedPartners(db *gorm.DB) {
	mustHaveDB(db)

	partners := getPartnerSeedData()
	inserted := 0
	updated := 0

	for _, p := range partners {
		var existing partner.Partner
		err := db.Where("external_id = ?", p.ExternalID).First(&existing).Error
		if err != nil {
			if err := db.Create(&p).Error; err != nil {
				log.Printf("Failed to create partner %s: %v", p.ExternalID, err)
			} else {
				inserted++
			}
		} else {
			p.Model = existing.Model
			if err := db.Save(&p).Error; err != nil {
				log.Printf("Failed to update partner %s: %v", p.ExternalID, err)
			} else {
				updated++
			}
		}
	}

	fmt.Printf("Partners seeding done: %d inserted, %d updated\n", inserted, updated)
}

func getPartnerSeedData() []partner.Partner {
	return []partner.Partner{
		{
			ExternalID:  "hyatt",
			Name:        "Hyatt Regency Yogyakarta",
			Description: "Hyatt Regency Yogyakarta adalah hotel bintang 5 dengan lokasi strategis di jantung kota. Menawarkan kenyamanan modern dengan sentuhan budaya Jawa serta hamparan taman tropis yang indah.",
			Category:    "Hotel",
			Location:    "Sleman, Yogyakarta",
			Address:     "Jl. Palagan Tentara Pelajar No.81, Sariharjo, Ngaglik, Sleman, Yogyakarta 55581",
			Image:       "https://images.unsplash.com/photo-1566073771259-6a8506099945?auto=format&fit=crop&w=600&q=80",
			Rating:      4.8,
			Price:       "IDR 1,450,000 / night",
			Phone:       "+62 274 1234 567",
			Website:     "https://hyatt.com/yogyakarta",
			Latitude:    -7.7361,
			Longitude:   110.3644,
		},
		{
			ExternalID:  "yudjum",
			Name:        "Gudeg Yu Djum",
			Description: "Kuliner legendaris paling terkenal di Yogyakarta yang menyajikan nasi gudeg manis gurih khas dengan telur bebek, krecek super pedas, dan ayam kampung empuk yang dimasak menggunakan resep turun-temurun di atas kayu bakar tradisional.",
			Category:    "Restaurant",
			Location:    "Kota Yogyakarta",
			Address:     "Jl. Wijilan No.167, Panembahan, Kraton, Kota Yogyakarta 55131",
			Image:       "https://images.unsplash.com/photo-1563379091339-03b21ab4a4f8?auto=format&fit=crop&w=600&q=80",
			Rating:      4.7,
			Price:       "IDR 25,000 – 60,000 / person",
			Phone:       "+62 274 515 965",
			Website:     "https://gudegyudjumwijilan167.com",
			Latitude:    -7.8057,
			Longitude:   110.3633,
		},
		{
			ExternalID:  "batikrumah",
			Name:        "Batik Rumah Jogja",
			Description: "Galeri batik tulis dan cap premium asli Yogyakarta. Menghadirkan motif klasik keraton hingga desain modern kontemporer yang elegan, dibuat dengan ketelitian seniman lokal terbaik.",
			Category:    "Souvenir Shop",
			Location:    "Bantul, Yogyakarta",
			Address:     "Jl. Imogiri Barat Km. 7, Sewon, Bantul, Yogyakarta 55188",
			Image:       "https://images.unsplash.com/photo-1524295988346-0b4b8e5e6f2f?auto=format&fit=crop&w=600&q=80",
			Rating:      4.9,
			Price:       "IDR 75,000 – 1,500,000 / item",
			Phone:       "+62 811-2345-678",
			Website:     "https://batikrumahjogja.co.id",
			Latitude:    -7.8618,
			Longitude:   110.3524,
		},
		{
			ExternalID:  "merapijeep",
			Name:        "Merapi Jeep Adventure",
			Description: "Sensasi petualangan menyusuri rute erupsi Gunung Merapi dengan mobil Jeep 4x4 klasik. Mengunjungi Bunker Kaliadem, Batu Alien, dan Museum Sisa Hartaku bersama pengemudi andal sekaligus pemandu cerita.",
			Category:    "Tour Guide",
			Location:    "Sleman, Yogyakarta",
			Address:     "Kaliurang, Hargobinangun, Pakem, Sleman, Yogyakarta 55582",
			Image:       "https://images.unsplash.com/photo-1533473359331-0135ef1b58bf?auto=format&fit=crop&w=600&q=80",
			Rating:      4.8,
			Price:       "IDR 350,000 – 650,000 / Jeep",
			Phone:       "+62 857-4321-0987",
			Website:     "https://merapijeepadventure.com",
			Latitude:    -7.6431,
			Longitude:   110.4455,
		},
		{
			ExternalID:  "jogjabike",
			Name:        "Jogja Bike Rental",
			Description: "Sewa sepeda harian, mingguan, maupun bulanan terbaik di Yogyakarta. Menyediakan sepeda perkotaan, sepeda gunung (MTB), hingga sepeda lipat berkualitas tinggi lengkap dengan helm dan gembok pengaman.",
			Category:    "Rental",
			Location:    "Kota Yogyakarta",
			Address:     "Jl. Sosrowijayan No.12, Gedong Tengengen, Kota Yogyakarta 55271",
			Image:       "https://images.unsplash.com/photo-1485965120184-e220f721d03e?auto=format&fit=crop&w=600&q=80",
			Rating:      4.6,
			Price:       "IDR 50,000 – 150,000 / day",
			Phone:       "+62 822-1111-2222",
			Website:     "https://jogjabikerental.com",
			Latitude:    -7.7931,
			Longitude:   110.3656,
		},
		{
			ExternalID:  "pakbudi",
			Name:        "Pak Budi Tour Guide",
			Description: "Pemandu wisata profesional bersertifikasi resmi HPI (Himpunan Pramuwisata Indonesia) dengan spesialisasi sejarah Keraton Yogyakarta, Candi Prambanan, dan Candi Borobudur serta rahasia tempat tersembunyi lokal.",
			Category:    "Tour Guide",
			Location:    "Kota Yogyakarta",
			Address:     "Kraton, Kota Yogyakarta 55132",
			Image:       "https://images.unsplash.com/photo-1519085360753-af0119f7cbe7?auto=format&fit=crop&w=600&q=80",
			Rating:      4.9,
			Price:       "IDR 300,000 – 600,000 / day",
			Phone:       "+62 813-9876-5432",
			Website:     "https://pakbudijogjaguide.com",
			Latitude:    -7.8054,
			Longitude:   110.3642,
		},
		{
			ExternalID:  "jogjatrans",
			Name:        "Jogja Trans Shuttle",
			Description: "Layanan sewa mobil dan mikrobus premium seperti Toyota Hiace, Alphard, dan Avanza. Menyediakan antar jemput bandara YIA serta paket wisata Jogja-Solo-Semarang dengan pengemudi profesional dan ramah.",
			Category:    "Rental",
			Location:    "Kota Yogyakarta",
			Address:     "Jl. Ringroad Utara No.45, Depok, Sleman, Yogyakarta 55281",
			Image:       "https://images.unsplash.com/photo-1544620347-c4fd4a3d5957?auto=format&fit=crop&w=600&q=80",
			Rating:      4.5,
			Price:       "IDR 400,000 – 1,200,000 / day",
			Phone:       "+62 812-4444-5555",
			Website:     "https://jogjatransshuttle.com",
			Latitude:    -7.7492,
			Longitude:   110.3867,
		},
		{
			ExternalID:  "kampoengsilver",
			Name:        "Kampoeng Silver Kotagede",
			Description: "Pusat kerajinan perak tertua di Yogyakarta. Kami menawarkan perhiasan perak ukir buatan tangan murni, miniatur candi perak, dan kelas membuat kerajinan perak sendiri bagi wisatawan asing maupun domestik.",
			Category:    "Souvenir Shop",
			Location:    "Kotagede, Yogyakarta",
			Address:     "Jl. Kemasan No.23, Kotagede, Kota Yogyakarta 55173",
			Image:       "https://images.unsplash.com/photo-1513519245088-0e12902e5a38?auto=format&fit=crop&w=600&q=80",
			Rating:      4.7,
			Price:       "IDR 150,000 – 5,000,000 / item",
			Phone:       "+62 274 375 123",
			Website:     "https://kampoengsilverkotagede.com",
			Latitude:    -7.8250,
			Longitude:   110.3980,
		},
	}
}
