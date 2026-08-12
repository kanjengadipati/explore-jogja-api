package seeds

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"pleco-api/internal/modules/destination"

	"gorm.io/gorm"
)

// normalizeReviewCount forces ReviewCount to equal len(Reviews) so the displayed
// review volume can never diverge from the actual seeded review array — a
// mismatch here is the most common cause of "4.9 (3,840 reviews)" appearing
// next to a 2-entry review list on the frontend. This also neutralizes the
// historical pattern of hand-typed, inaccurate ReviewCount values.
func normalizeReviewCount(d *destination.Destination) {
	if d == nil {
		return
	}
	d.ReviewCount = len(d.Reviews)
}


func SeedDestinations(db *gorm.DB) {
	mustHaveDB(db)

	var count int64
	db.Model(&destination.Destination{}).Count(&count)

	// Attempt to load from JSON first
	jsonPath := "internal/seeds/destinations.json"
	data, err := os.ReadFile(jsonPath)
	if err == nil {
		var dests []destination.Destination
		if err := json.Unmarshal(data, &dests); err == nil {
			fmt.Printf("Found %d destinations in JSON seed file. Syncing to database...\n", len(dests))
			inserted := 0
			updated := 0
			for _, d := range dests {
				normalizeReviewCount(&d)
				var existing destination.Destination
				err := db.Where("external_id = ?", d.ExternalID).First(&existing).Error
				if err != nil {
					// Does not exist, create it
					if err := db.Create(&d).Error; err != nil {
						log.Printf("Failed to create destination %s: %v", d.ExternalID, err)
					} else {
						inserted++
					}
				} else {
					// Exists, update it
					d.ID = existing.ID
					d.CreatedAt = existing.CreatedAt
					if err := db.Save(&d).Error; err != nil {
						log.Printf("Failed to update destination %s: %v", d.ExternalID, err)
					} else {
						updated++
					}
				}
			}
			fmt.Printf("Destinations sync done: %d inserted, %d updated\n", inserted, updated)
			return
		} else {
			log.Printf("Failed to unmarshal destinations JSON: %v", err)
		}
	} else {
		log.Printf("Could not read destinations.json: %v. Falling back to hardcoded seed data.", err)
	}

	// Fallback to hardcoded seed data if JSON not present or failed to parse
	if count > 0 {
		fmt.Println("Destinations already seeded, skipping")
		return
	}

	dests := getDestinationSeedData()
	for i := range dests {
		normalizeReviewCount(&dests[i])
	}
	if err := db.CreateInBatches(dests, 10).Error; err != nil {
		log.Printf("Failed to seed destinations: %v", err)
		return
	}
	fmt.Printf("Destinations seeding done (%d records)\n", len(dests))
}

func ja(v interface{}) destination.JSONArr {
	b, _ := json.Marshal(v)
	var arr destination.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func jm(v interface{}) destination.JSONMap {
	b, _ := json.Marshal(v)
	var m destination.JSONMap
	json.Unmarshal(b, &m)
	return m
}

func getDestinationSeedData() []destination.Destination {
	return []destination.Destination{
		// ==========================================
		// 1. PRAMBANAN TEMPLE
		// ==========================================
		{
			ExternalID:   "prambanan",
			Name:         "Prambanan Temple",
			Tagline:      "The Pinnacle of Royal Hindu Architecture",
			Category:     "heritage",
			Location:     "Sleman, Yogyakarta",
			SubRegion:    "Sleman",
			Images:       ja([]string{"https://images.unsplash.com/photo-1578469550956-0e16b69c6a3d?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1621869606578-1561708a7e09?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1566559631133-969041fc5583?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.8,
			ReviewCount:  3840,
			Description:  "Prambanan is the largest Hindu temple compound in Indonesia, built in the 9th century under the Sanjaya dynasty of the Mataram Kingdom. Dedicated to the Trimurti — Shiva, Vishnu, and Brahma — the main temple soars to 47 meters, flanked by 240 smaller temples across 40 hectares. The compound became a UNESCO World Heritage Site in 1991.",
			Story:        "Legend tells of Roro Jonggrang, a princess who demanded Prince Bandung Bondowoso build 1,000 temples in one night. He nearly succeeded with supernatural help, but Roro Jonggrang tricked the roosters into crowing early. Enraged by the deception, Bandung cursed her into the final 1,000th stone statue — which remains in the Shiva chamber to this day.",
			TicketPrice:  "IDR 375,000 (Foreign) / IDR 50,000 (Domestic)",
			OpeningHours: "06:30 AM – 05:00 PM Daily (Sunrise package: 05:30 AM)",
			Facilities:   ja([]string{"Visitor Information Center", "Audio Guide Rental", "Wheelchair Access", "Large Parking Area", "Batik Souvenir Arcades", "Traditional Restaurants", "Clean Restrooms", "Ramayana Ballet Stage"}),
			TravelTips:   ja([]string{"Arrive around 03:30 PM for golden hour light through the spires.", "Wear modest clothing; sarongs are provided at the entrance.", "Catch the Ramayana Ballet on Tuesday, Thursday, and Saturday evenings during dry season.", "Buy the Prambanan + Ratu Boko combo ticket for best value."}),
			BestTime:     "May to October (Dry Season), late afternoon for sunset photography",
			Weather:      jm(map[string]string{"temp": "28°C", "condition": "Sunny", "status": "Perfect weather for exploring Prambanan today."}),
			Latitude:     -7.7520,
			Longitude:    110.4914,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "An absolutely magical experience. The architecture is breathtaking and the Ramayana Ballet was unforgettable."},
				{"id": "r2", "userName": "Yuki Tanaka", "userAvatar": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150", "rating": 4.8, "date": "2026-07-02", "comment": "Watching the sunset behind the spires is a lifetime memory. Hire a local guide for deeper insights."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "p-p1", "name": "The Phoenix Hotel Yogyakarta", "category": "hotel", "image": "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600", "rating": 4.9, "price": "IDR 1,500,000 / night", "distance": "14 km from Prambanan", "description": "A luxurious colonial heritage hotel with royal Javanese spa and grand courtyard pool.", "address": "Jl. Jend. Sudirman No.9, Yogyakarta"},
				{"id": "p-p2", "name": "Abhayagiri Restaurant", "category": "restaurant", "image": "https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?q=80&w=600", "rating": 4.8, "price": "IDR 150,000 – 300,000 / person", "distance": "3.2 km", "description": "Hilltop fine dining with panoramic sunset views of Prambanan and Mount Merapi.", "address": "Samberwatu, Sambirejo, Prambanan"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is there a dress code?", "a": "Yes. Modest wear covering knees and shoulders is recommended. Sarongs are provided."},
				{"q": "When does the Ramayana Ballet perform?", "a": "Tuesday, Thursday, and Saturday evenings during dry season (May–October)."},
				{"q": "Can I visit for sunrise?", "a": "Yes. The Mruput Prambanan package opens at 05:30 AM."},
			}),
		},

		// ==========================================
		// 2. MALIOBORO STREET
		// ==========================================
		{
			ExternalID:   "malioboro",
			Name:         "Malioboro Street",
			Tagline:      "The Soul and Lifeline of Yogyakarta",
			Category:     "culinary",
			Location:     "Yogyakarta City",
			SubRegion:    "Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1543874768-af0b9c4090d5?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1617591387509-2fcba0c80c42?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1581456495146-65a71b2c8e52?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.6,
			ReviewCount:  9280,
			Description:  "Jalan Malioboro is Yogyakarta's most iconic street — a 2.5 km artery stretching from Tugu Station to the Sultan's Palace (Kraton). Lined with colonial-era buildings, hundreds of batik shops, street food vendors, and alive with the sounds of street musicians and the clatter of traditional horse-drawn carriages (Andong).",
			Story:        "Malioboro sits on the sacred cosmological axis (Philosophical Axis of Yogyakarta) connecting Mount Merapi in the north, the Kraton in the center, and the mystical South Sea (Parangtritis) in the south. This axis represents the Javanese belief in cosmic balance between the spiritual and physical worlds.",
			TicketPrice:  "Free Entry",
			OpeningHours: "24 Hours (Best experienced evening 06:00 PM – 11:00 PM)",
			Facilities:   ja([]string{"Pedestrian Walkways", "Batik & Craft Stores", "Street Food Stalls", "Live Street Musicians", "Becak & Andong Stands", "Historic Colonial Buildings", "Tourist Police Post", "ATMs & Money Changers"}),
			TravelTips:   ja([]string{"Start from Tugu Station and walk south to the Kraton for the full experience.", "Try Gudeg Yu Djum — the most famous gudeg restaurant in Jogja.", "Bargain with a smile when buying batik; it's part of the culture.", "Visit on Tuesday night for the weekly Malioboro Street Festival."}),
			BestTime:     "Every evening after 06:00 PM, especially during full moon celebrations",
			Weather:      jm(map[string]string{"temp": "27°C", "condition": "Clear Evening", "status": "Perfect cool breeze for an evening walk in Malioboro."}),
			Latitude:     -7.7928,
			Longitude:    110.3658,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "The energy of Malioboro at night is electric. Street food, music, and the friendliest people."},
				{"id": "r2", "userName": "Budi Santoso", "userAvatar": "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?q=80&w=150", "rating": 4.5, "date": "2026-07-10", "comment": "Best batik shopping in all of Indonesia. Don't forget to try the bakpia pathok!"},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "m-p1", "name": "Grand Inna Malioboro", "category": "hotel", "image": "https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=600", "rating": 4.7, "price": "IDR 1,100,000 / night", "distance": "0 km (on Malioboro)", "description": "Iconic heritage hotel established in 1908, right on Jalan Malioboro.", "address": "Jl. Malioboro No.60, Yogyakarta"},
				{"id": "m-p2", "name": "Gudeg Yu Djum", "category": "restaurant", "image": "https://images.unsplash.com/photo-1504674900247-0877df9cc836?q=80&w=600", "rating": 4.8, "price": "IDR 25,000 – 60,000 / person", "distance": "1.2 km from Malioboro", "description": "The most legendary gudeg restaurant in Yogyakarta since 1950.", "address": "Jl. Wijaya Kusuma No.20, Yogyakarta"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is Malioboro wheelchair accessible?", "a": "Yes. The pedestrian walkways are wide and well-maintained."},
				{"q": "What is the best way to get there?", "a": "Arrive by train at Tugu Station, which is at the northern end of Malioboro."},
				{"q": "Is it safe at night?", "a": "Yes. Malioboro is well-lit and patrolled by tourist police 24/7."},
			}),
		},

		// ==========================================
		// 3. PARANGTRITIS BEACH
		// ==========================================
		{
			ExternalID:   "parangtritis",
			Name:         "Parangtritis Beach",
			Tagline:      "Mystical Golden Sands of the Southern Realm",
			Category:     "beach",
			Location:     "Bantul, Yogyakarta",
			SubRegion:    "Bantul",
			Images:       ja([]string{"https://images.unsplash.com/photo-1602137704924-9a038cfb5253?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1519046904884-53103b34b206?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.7,
			ReviewCount:  4230,
			Description:  "Framed by dramatic black volcanic sands, towering karst cliffs, and roaring waves from the Indian Ocean, Parangtritis is Yogyakarta's most legendary beach. The vast coastline stretches for kilometers, backed by massive sand dunes (Gumuk Pasir) perfect for sandboarding.",
			Story:        "Deeply woven with Javanese cosmology, Parangtritis is believed to be the sacred gateway to the undersea palace of Kanjeng Ratu Kidul — the mystical Queen of the Southern Seas. Sultan Hamengkubuwono I reportedly discovered the beach while hunting, and the royal family has maintained a spiritual connection to this place ever since.",
			TicketPrice:  "IDR 15,000 / person",
			OpeningHours: "24 Hours Daily",
			Facilities:   ja([]string{"Horse-Drawn Carriage Rides", "ATV Rentals", "Sandboarding at Gumuk Pasir", "Fresh Coconut Stalls", "Cliffside Gazebos", "Lifeguard Posts", "Local Seafood Warungs", "Paragliding (nearby)"}),
			TravelTips:   ja([]string{"Do NOT swim — the undercurrents are extremely dangerous.", "Rent a horse carriage (Andong) for a sunset ride along the tideline.", "Head to Gumuk Pasir (2 km west) for world-class sandboarding.", "Visit during full moon for the sacred Labuhan ceremony."}),
			BestTime:     "June to August for crisp sunsets; full moon for cultural ceremonies",
			Weather:      jm(map[string]string{"temp": "29°C", "condition": "Ocean Breeze", "status": "Beautiful clear skies over the South Sea today."}),
			Latitude:     -8.0253,
			Longitude:    110.3298,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "The sandboarding at Gumuk Pasir was incredible! The black volcanic sand is so unique."},
				{"id": "r2", "userName": "Yuki Tanaka", "userAvatar": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150", "rating": 4.6, "date": "2026-07-02", "comment": "The sunset here is otherworldly. The mystical energy of this place is palpable."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "pt-p1", "name": "Queen of the South Resort", "category": "hotel", "image": "https://images.unsplash.com/photo-1571896349842-33c89424de2d?q=80&w=600", "rating": 4.8, "price": "IDR 1,400,000 / night", "distance": "1.5 km from beach", "description": "Clifftop resort with infinity pool overlooking the Indian Ocean.", "address": "Parangrejo, Purwosari, Bantul"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Can I swim in the sea?", "a": "Strictly prohibited. The undercurrents are extremely powerful."},
				{"q": "Is there sandboarding?", "a": "Yes! At Gumuk Pasir Parangkusumo, 2 km west of Parangtritis."},
			}),
		},

		// ==========================================
		// 4. MOUNT MERAPI LAVA TOUR
		// ==========================================
		{
			ExternalID:   "merapi",
			Name:         "Mount Merapi Lava Tour",
			Tagline:      "An Unforgettable Offroad Journey on an Active Volcano",
			Category:     "adventure",
			Location:     "Sleman, Yogyakarta",
			SubRegion:    "Sleman",
			Images:       ja([]string{"https://images.unsplash.com/photo-1655861467672-215aaeb119ff?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1704287824391-b3bc1e04ea33?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1518070588484-2b53926cba76?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.8,
			ReviewCount:  5120,
			Description:  "Ride inside open-cabin vintage 4x4 Willys Jeeps along the trails left by Mount Merapi's historical eruptions. The Lava Tour takes you through devastated villages, buried homes, and the eerie Kaliadem Bunker — all with the smoking crater as your backdrop.",
			Story:        "Mount Merapi (2,930m) is one of the world's most active volcanoes, erupting regularly since 1548. The devastating 2010 eruption killed 353 people and buried entire villages. The local Javanese hold deep spiritual reverence for the mountain, believing it is home to supernatural guardians. Every year, the Labuhan ceremony is held to offer gifts to the mountain spirits.",
			TicketPrice:  "IDR 350,000 – 650,000 per Jeep (2-4 persons)",
			OpeningHours: "04:30 AM – 06:00 Daily (Sunrise Tour departs 04:30 AM)",
			Facilities:   ja([]string{"Willys Jeep Fleet", "Certified Offroad Drivers", "Protective Helmets & Masks", "Merapi Lava Museum", "Kaliadem Bunker", "Viewpoint Cafes", "Local Guide Services"}),
			TravelTips:   ja([]string{"Book the Sunrise Jeep Tour — watching dawn break over an active volcano is unforgettable.", "Bring a windbreaker; it's significantly cooler at higher elevations.", "Wear the provided face mask on dusty trail sections.", "Combine with a visit to Selo village for traditional Javanese lunch."}),
			BestTime:     "May to September (Dry Season), sunrise tours are best",
			Weather:      jm(map[string]string{"temp": "22°C", "condition": "Mist & Sunrise", "status": "Stunning cloudless view of the volcano summit today."}),
			Latitude:     -7.5960,
			Longitude:    110.4463,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "The sunrise jeep tour was the highlight of our Jogja trip. Absolutely thrilling!"},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "me-p1", "name": "Kopi Klotok Pakem", "category": "restaurant", "image": "https://images.unsplash.com/photo-1501339847302-ac426a4a7cbb?q=80&w=600", "rating": 4.9, "price": "IDR 20,000 – 45,000 / person", "distance": "6.5 km from Tour Base", "description": "Rustic Javanese village cafe serving traditional kopi klotok and nasi campur.", "address": "Jl. Kaliurang Km 16, Pakem, Sleman"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is it safe for kids?", "a": "Yes. Families can choose the Short Route (1.5 hours)."},
				{"q": "What is inside the Kaliadem Bunker?", "a": "A steel-reinforced shelter built to withstand volcanic heat. It saved two people in 2006."},
			}),
		},

		// ==========================================
		// 5. TAMAN SARI WATER CASTLE
		// ==========================================
		{
			ExternalID:   "tamansari",
			Name:         "Taman Sari Water Castle",
			Tagline:      "The Secret Royal Bathing Pools of the Sultanate",
			Category:     "heritage",
			Location:     "Yogyakarta City",
			SubRegion:    "Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1625506276715-76ad63823181?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1596402184320-417e7178b2cd?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1584810359583-96fc3448beaa?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.6,
			ReviewCount:  3120,
			Description:  "Built in the mid-18th century as a private pleasure park for Sultan Hamengkubuwono I, Taman Sari is a stunning architectural mixture of Javanese, Portuguese, and European styles. The complex features turquoise bathing pools, an underground mosque (Sumur Gumuling), and a labyrinth of pastel-painted houses.",
			Story:        "Sultan Hamengkubuwono I built Taman Sari as a resting palace, defense castle, and mystical meditation sanctuary. The name means 'Beautiful Garden.' After a Portuguese architect fell in love with the Sultan's daughter, he was forced to build this masterpiece as a dowry. The bathing pools were designed so the Sultan could watch his queens bathe from above through carved stone windows.",
			TicketPrice:  "IDR 15,000 / person (Guide recommended: IDR 50,000)",
			OpeningHours: "09:00 AM – 03:30 PM Daily",
			Facilities:   ja([]string{"English-Speaking Royal Guides", "Turquoise Bathing Pools", "Underground Mosque (Sumur Gumuling)", "Heritage Photo Spots", "Kampung Cyber Nearby", "Artisan Craft Galleries", "Clean Restrooms"}),
			TravelTips:   ja([]string{"Hire a local guide to navigate the labyrinth — they know the best photo spots.", "The underground Sumur Gumuling mosque is a must-visit for its unique acoustics.", "Visit Kampung Cyber (Cyber Village) next door for colorful street art.", "Morning light (09:30–11:30 AM) is best for photography."}),
			BestTime:     "Morning 09:30 AM – 11:30 AM for best light and fewer crowds",
			Weather:      jm(map[string]string{"temp": "28°C", "condition": "Partly Cloudy", "status": "Warm sunshine, perfect for exploring the outdoor pools."}),
			Latitude:     -7.8101,
			Longitude:    110.3592,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "The turquoise pools are stunning! Our guide made the history come alive."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "ts-p1", "name": "The Phoenix Hotel Yogyakarta", "category": "hotel", "image": "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600", "rating": 4.9, "price": "IDR 1,500,000 / night", "distance": "4.2 km", "description": "Colonial elegance blended with royal Javanese architecture.", "address": "Jl. Jend. Sudirman No.9, Yogyakarta"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is Sumur Gumuling open?", "a": "Yes, it's accessible via a narrow staircase inside the complex."},
				{"q": "Can we photograph the pools?", "a": "Yes! Personal photography is free and encouraged."},
			}),
		},

		// ==========================================
		// 6. GOA JOMBLANG CAVE
		// ==========================================
		{
			ExternalID:   "goajomblang",
			Name:         "Goa Jomblang Cave",
			Tagline:      "The Celestial Beam of Heavenly Light",
			Category:     "nature",
			Location:     "Gunungkidul, Yogyakarta",
			SubRegion:    "Gunungkidul",
			Images:       ja([]string{"https://images.unsplash.com/photo-1628047563315-d1e8b8d222b9?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1571607023618-c93917452ed3?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1519046904884-53103b34b206?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.9,
			ReviewCount:  1840,
			Description:  "Goa Jomblang is a vertical collapse-sinkhole cave, 60 meters deep. You rappel down into a mystical underground ancient forest. At noon, a single beam of sunlight pierces through the cave ceiling, illuminating the forest floor in what locals call the 'Light of Heaven.'",
			Story:        "Formed hundreds of thousands of years ago when the ground collapsed into a vertical limestone sinkhole. The cave was first explored in 1984 by a group of spelunkers from the Yogyakarta Adventurer Club. The light beam phenomenon occurs because the sinkhole faces almost directly upward, allowing sunlight to penetrate at midday.",
			TicketPrice:  "IDR 500,000 / person (includes all equipment and guide)",
			OpeningHours: "07:30 AM – 12:30 PM (Light beam visible 11:15 AM – 12:30 PM)",
			Facilities:   ja([]string{"Certified SRT Guides", "Full Rappelling Equipment", "Rubber Boots Provided", "Outdoor Shower Rooms", "Traditional Lunch Box", "Resting Gazebos", "First Aid Station"}),
			TravelTips:   ja([]string{"Advance bookings are MANDATORY (80 people/day limit).", "Arrive early to secure the 11:15 AM light beam slot.", "Bring a dry change of clothes and a towel.", "Pack a headlamp for exploring side tunnels.", "Combine with Goa Grubug for a longer caving adventure."}),
			BestTime:     "Sunny days between 11:15 AM and 12:30 PM for the heavenly light beam",
			Weather:      jm(map[string]string{"temp": "27°C", "condition": "Sunny Outside", "status": "Perfect clear skies — the light beam will be spectacular today!"}),
			Latitude:     -8.0287,
			Longitude:    110.6384,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "Rappelling into the cave was thrilling, and the light beam at noon was truly heavenly."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "gj-p1", "name": "Gunungkidul Cave Guides Guild", "category": "guide", "image": "https://images.unsplash.com/photo-1544717297-fa95b6ee9643?q=80&w=600", "rating": 4.9, "price": "Included in ticket", "distance": "On-Site", "description": "Highly trained SRT safety experts with decades of cave experience.", "address": "Pacarejo, Semanu, Gunungkidul"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Do I need caving experience?", "a": "No. Full training and equipment are provided."},
				{"q": "Are there age restrictions?", "a": "Recommended for ages 10–60. Children under 10 are not permitted."},
			}),
		},

		// ==========================================
		// 7. KALIBIRU NATIONAL PARK
		// ==========================================
		{
			ExternalID:   "kalibiru",
			Name:         "Kalibiru National Park",
			Tagline:      "Highland Serenity and Panoramic Reservoir Views",
			Category:     "nature",
			Location:     "Kulon Progo, Yogyakarta",
			SubRegion:    "Kulon Progo",
			Images:       ja([]string{"https://images.unsplash.com/photo-1571607023618-c93917452ed3?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1441974231531-c6227db76b6e?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1470071459604-3b5ec3a7fe05?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.7,
			ReviewCount:  2180,
			Description:  "Located in the Menoreh Hills, Kalibiru National Park offers breathtaking panoramic views of the Sermo Reservoir surrounded by lush green pine forests. Famous for its treetop viewing platforms, flying fox ziplines, and eco-tourism model that transformed degraded forest into a thriving park.",
			Story:        "Kalibiru was originally a dry, degraded forest area. Through collaborative reforestation efforts by the local community and government, it blossomed into a pioneering eco-tourism model that has won multiple environmental awards. The park now protects over 400 hectares of regenerated tropical forest.",
			TicketPrice:  "IDR 20,000 / person (additional fees for zipline activities)",
			OpeningHours: "08:00 AM – 05:00 PM Daily",
			Facilities:   ja([]string{"Treetop Viewing Platforms", "Flying Fox Ziplines", "Trekking Trails", "Local Viewpoint Cafes", "Spacious Parking", "Clean Restrooms", "Camping Grounds", "Photography Platforms"}),
			TravelTips:   ja([]string{"Arrive early (before 09:00 AM) for the clearest mountain views.", "Wear comfortable hiking shoes for the trails.", "Hire a local photographer at the platforms for stunning photos.", "Combine with a visit to Waduk Sermo (Sermo Reservoir) nearby."}),
			BestTime:     "June to September (Dry Season) for clearest skies",
			Weather:      jm(map[string]string{"temp": "24°C", "condition": "Clear & Cool", "status": "Beautiful clear skies with a cool mountain breeze."}),
			Latitude:     -7.8052,
			Longitude:    110.1293,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "The view from the treetop platform over the reservoir is breathtaking. Perfect for photos!"},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "kb-p1", "name": "Sermo Lakeside Camping", "category": "rental", "image": "https://images.unsplash.com/photo-1504280390367-361c6d9f38f4?q=80&w=600", "rating": 4.7, "price": "IDR 150,000 / night", "distance": "4.5 km", "description": "Camp under the stars on the edge of Sermo Reservoir.", "address": "Sermo Reservoir Area, Kulon Progo"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is it safe to climb the platforms?", "a": "Yes. Every platform is equipped with professional safety harnesses."},
				{"q": "Can I camp overnight?", "a": "Yes. Designated camping grounds are available near the reservoir."},
			}),
		},

		// ==========================================
		// 8. KERATON YOGYAKARTA (NEW)
		// ==========================================
		{
			ExternalID:   "keraton",
			Name:         "Keraton Yogyakarta",
			Tagline:      "The Living Heart of the Sultanate",
			Category:     "heritage",
			Location:     "Yogyakarta City",
			SubRegion:    "Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1596402184320-417e7178b2cd?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1581456495146-65a71b2c8e52?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1584810359583-96fc3448beaa?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.7,
			ReviewCount:  4500,
			Description:  "Keraton Ngayogyakarta Hadiningrat is the official palace of the Sultan of Yogyakarta and a living museum of Javanese royal culture. Built in 1755 by Sultan Hamengkubuwono I, the palace complex features stunning Javanese architecture, sacred gamelan instruments, royal artifacts, and the Sultan's personal guard (Abdi Dalem).",
			Story:        "The Kraton was built on the Philosophical Axis of Yogyakarta — the spiritual line connecting Mount Merapi, the Kraton, and the South Sea. According to Javanese cosmology, this axis represents the connection between the divine realm, the human world, and the spirit sea. The palace is still actively used by the Sultan and his court.",
			TicketPrice:  "IDR 15,000 / person (additional IDR 5,000 for camera)",
			OpeningHours: "08:30 AM – 02:00 PM Tuesday–Thursday & Saturday–Sunday",
			Facilities:   ja([]string{"Royal Museum", "Gamelan Performance Hall", "Sultan's Personal Guard", "Traditional Javanese Architecture", "Royal Gardens", "Gift Shop", "Clean Restrooms"}),
			TravelTips:   ja([]string{"Visit during the gamelan performance (usually 10:00 AM and 12:00 PM).", "Wear modest clothing — this is still an active royal palace.", "Hire a guide to understand the deep symbolism of each room.", "Don't miss the Bangsal Kencana (Golden Pavilion)."}),
			BestTime:     "Tuesday to Sunday mornings for gamelan performances",
			Weather:      jm(map[string]string{"temp": "28°C", "condition": "Partly Cloudy", "status": "Perfect weather for exploring the royal palace."}),
			Latitude:     -7.8054,
			Longitude:    110.3642,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Budi Santoso", "userAvatar": "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?q=80&w=150", "rating": 5, "date": "2026-07-10", "comment": "A must-visit to understand Javanese royal culture. The gamelan performance was mesmerizing."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "k-p1", "name": "Royal Ambarrukmo Hotel", "category": "hotel", "image": "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600", "rating": 4.8, "price": "IDR 1,200,000 / night", "distance": "3.5 km from Kraton", "description": "Heritage luxury hotel once used by visiting royalty.", "address": "Jl. Babarsari No.53, Yogyakarta"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is the palace still inhabited?", "a": "Yes. The Sultan's family still uses part of the complex."},
				{"q": "What days is it open?", "a": "Tuesday through Sunday, 08:30 AM – 02:00 PM. Closed on Monday."},
			}),
		},

		// ==========================================
		// 9. RATU BOKO PALACE (NEW)
		// ==========================================
		{
			ExternalID:   "ratuboko",
			Name:         "Ratu Boko Palace",
			Tagline:      "The Lost Kingdom on the Hilltop",
			Category:     "heritage",
			Location:     "Sleman, Yogyakarta",
			SubRegion:    "Sleman",
			Images:       ja([]string{"https://images.unsplash.com/photo-1782392487647-7fee9715d1f5?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.7,
			ReviewCount:  3200,
			Description:  "Ratu Boko is an archaeological palace complex perched on a plateau 3 km south of Prambanan. Unlike other Javanese sites which are temples, Ratu Boko displays attributes of a fortified royal settlement — complete with defensive walls, dry moats, bathing pools, and meditation chambers. It is most famous for its breathtaking sunset views.",
			Story:        "The original name remains unclear, but local inhabitants named the site after King Boko, the legendary king from the Roro Jonggrang folklore. Archaeologists believe it was a palace complex of the Sailendra or Mataram Kingdom. The site covers 16 hectares across two hamlets and offers one of the most dramatic sunset viewpoints in all of Java.",
			TicketPrice:  "IDR 40,000 (Adult) / IDR 20,000 (Child) / IDR 10,000 (Parking)",
			OpeningHours: "06:00 AM – 05:00 PM Daily (Sunset package available)",
			Facilities:   ja([]string{"Shuttle Bus from Prambanan", "Sunset Viewpoint Platform", "Ancient Stone Gate (Gapura)", "Bathing Pools Ruins", "Meditation Caves", "Restaurant & Cafe", "Parking Area", "Gift Shop"}),
			TravelTips:   ja([]string{"Arrive 2 hours before sunset to explore and secure the best spot.", "Buy the Prambanan + Ratu Boko combo ticket with included shuttle.", "Bring a picnic mat to relax on the grass while waiting for sunset.", "Comfortable walking shoes are essential — the complex is large."}),
			BestTime:     "Late afternoon for sunset; dry season (May–September) for clearest skies",
			Weather:      jm(map[string]string{"temp": "27°C", "condition": "Clear", "status": "Perfect conditions for a spectacular sunset at Ratu Boko."}),
			Latitude:     -7.7700,
			Longitude:    110.4830,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "The sunset from the ancient ruins was the most magical moment of our trip."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "rb-p1", "name": "Boko Sunset Resto", "category": "restaurant", "image": "https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?q=80&w=600", "rating": 4.4, "price": "IDR 80,000 – 200,000 / person", "distance": "On-Site", "description": "Dine with panoramic views of the sunset over the Prambanan plain.", "address": "Jl. Ratu Boko, Bokoharjo, Prambanan"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is it a temple?", "a": "No. Ratu Boko is an ancient palace complex, not a temple."},
				{"q": "Can I visit for sunset?", "a": "Yes! Sunset packages are available. Arrive 2 hours before sunset."},
			}),
		},

		// ==========================================
		// 10. TIMANG BEACH (NEW)
		// ==========================================
		{
			ExternalID:   "timang",
			Name:         "Timang Beach",
			Tagline:      "The Thrilling Rope Bridge Over Wild Waves",
			Category:     "adventure",
			Location:     "Gunungkidul, Yogyakarta",
			SubRegion:    "Gunungkidul",
			Images:       ja([]string{"https://images.unsplash.com/photo-1595323264810-3dd19b1be56f?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1519046904884-53103b34b206?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.8,
			ReviewCount:  2800,
			Description:  "Timang Beach is a dramatic limestone cliff beach famous for its traditional wooden rope bridge (Jembatan Timang) that spans across crashing Indian Ocean waves. Originally built by lobster fishermen, the bridge has become one of Yogyakarta's most thrilling adventure attractions.",
			Story:        "For generations, local fishermen built and rebuilt this precarious rope bridge to reach the lobster-rich rocks on the other side of the cove. The bridge is made entirely from wooden planks and nylon ropes, hand-pulled by local volunteers. The tradition of lobster fishing here dates back centuries, and the bridge has been featured on the Japanese TV show 'Legends of the Hidden Temple.'",
			TicketPrice:  "IDR 100,000 – 200,000 / person (rope bridge crossing)",
			OpeningHours: "07:00 AM – 05:00 PM Daily",
			Facilities:   ja([]string{"Rope Bridge Crossing", "Traditional Lobster Dining", "Cliffside Viewpoints", "Local Guide Services", "Parking Area", "Warung (Food Stalls)", "ATV Rental (nearby)"}),
			TravelTips:   ja([]string{"The rope bridge crossing is not for the faint of heart — hold on tight!", "Wear closed-toe shoes with good grip.", "Try the fresh grilled lobster — it's the specialty here.", "Visit on a weekday to avoid long queues for the bridge."}),
			BestTime:     "May to September for calmer seas; morning for best light",
			Weather:      jm(map[string]string{"temp": "28°C", "condition": "Sunny", "status": "Clear skies — perfect for the rope bridge adventure!"}),
			Latitude:     -8.1350,
			Longitude:    110.6580,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Budi Santoso", "userAvatar": "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?q=80&w=150", "rating": 5, "date": "2026-07-10", "comment": "Walking across the rope bridge with waves crashing below was terrifying and amazing!"},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "t-p1", "name": "Timang Lobster Warung", "category": "restaurant", "image": "https://images.unsplash.com/photo-1504674900247-0877df9cc836?q=80&w=600", "rating": 4.7, "price": "IDR 100,000 – 250,000 / person", "distance": "On-Site", "description": "Fresh grilled lobster caught daily by local fishermen.", "address": "Timang Beach, Tepus, Gunungkidul"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is the rope bridge safe?", "a": "Yes. It is inspected and maintained daily by local fishermen."},
				{"q": "Can children cross the bridge?", "a": "Children under 12 are not recommended to cross."},
			}),
		},

		// ==========================================
		// 11. TEBING BREKSI (NEW)
		// ==========================================
		{
			ExternalID:   "tebingbreksi",
			Name:         "Tebing Breksi",
			Tagline:      "The Dramatic Limestone Cliff of Yogyakarta",
			Category:     "nature",
			Location:     "Sleman, Yogyakarta",
			SubRegion:    "Sleman",
			Images:       ja([]string{"https://images.unsplash.com/photo-1524445361389-3d2c21109015?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1441974231531-c6227db76b6e?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.6,
			ReviewCount:  1900,
			Description:  "Tebing Breksi is a former limestone quarry transformed into one of Yogyakarta's most spectacular viewpoints. The dramatic carved cliffs feature layered geological formations dating back millions of years. At the summit, you can see panoramic views stretching from Prambanan to the ocean.",
			Story:        "Originally an active limestone quarry, Tebing Breksi was closed by the government in 2014 to preserve the geological heritage. Local communities then developed it into a tourist attraction with carved pathways, amphitheater, and viewing platforms. The cliff face reveals geological layers formed over millions of years, making it both a scenic and scientific site.",
			TicketPrice:  "IDR 10,000 / person",
			OpeningHours: "06:00 AM – 06:00 PM Daily",
			Facilities:   ja([]string{"Carved Stone Pathways", "Summit Viewing Platform", "Open-Air Amphitheater", "Sunrise & Sunset Spots", "Food Stalls", "Parking Area", "Restrooms"}),
			TravelTips:   ja([]string{"Visit at sunrise (05:30 AM) for golden light over Prambanan.", "The amphitheater sometimes hosts evening cultural performances.", "Combine with a visit to nearby Candi Ijo for a full day trip.", "Wear sturdy shoes — the stone paths can be slippery."}),
			BestTime:     "Sunrise (05:30 AM) or sunset (05:00 PM) for dramatic lighting",
			Weather:      jm(map[string]string{"temp": "26°C", "condition": "Clear", "status": "Perfect visibility for panoramic views from the cliff summit."}),
			Latitude:     -7.7710,
			Longitude:    110.4750,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Yuki Tanaka", "userAvatar": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150", "rating": 4.5, "date": "2026-07-02", "comment": "The geological formations are stunning. Great spot for photography enthusiasts."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "tb-p1", "name": "Candi Ijo Restaurant", "category": "restaurant", "image": "https://images.unsplash.com/photo-1504674900247-0877df9cc836?q=80&w=600", "rating": 4.5, "price": "IDR 30,000 – 80,000 / person", "distance": "2 km", "description": "Traditional Javanese restaurant near Candi Ijo with hilltop views.", "address": "Sambirejo, Prambanan, Sleman"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is it safe to climb?", "a": "Yes. Carved pathways and safety railings are installed."},
				{"q": "Is there an entrance fee?", "a": "Yes, IDR 10,000 per person. Parking IDR 5,000."},
			}),
		},

		// ==========================================
		// 12. PINDUL CAVE (NEW)
		// ==========================================
		{
			ExternalID:   "pindul",
			Name:         "Pindul Cave tubing",
			Tagline:      "Glide Through Underground Rivers in Ancient Caves",
			Category:     "adventure",
			Location:     "Gunungkidul, Yogyakarta",
			SubRegion:    "Gunungkidul",
			Images:       ja([]string{"https://images.unsplash.com/photo-1550075099-dcd6d1513b87?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1511497584788-876760111969?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.7,
			ReviewCount:  2400,
			Description:  "Pindul Cave tubing is an underground river tubing adventure through a 350-meter limestone cave. You float on inner tubes through crystal-clear underground water, passing by stunning stalactites and stalagmites illuminated by guide-held lights. The experience takes about 45 minutes.",
			Story:        "Pindul Cave was discovered by local explorers in 2008. The underground river was found to be deep enough (5–12 meters) and calm enough for safe tubing. The cave ceiling is adorned with ancient stalactites formed over millions of years, some reaching down to nearly touch the water surface.",
			TicketPrice:  "IDR 100,000 / person (includes tube, guide, and life jacket)",
			OpeningHours: "08:00 AM – 04:00 PM Daily",
			Facilities:   ja([]string{"Full Tubing Equipment", "Certified Cave Guides", "Life Jackets", "Underwater Flashlights", "Changing Rooms", "Freshwater Showers", "Parking Area", "Warung (Food Stalls)"}),
			TravelTips:   ja([]string{"Wear swimwear under your clothes — you will get wet!", "Waterproof cameras or phone cases are essential.", "Combine with a visit to nearby Bejiharjo Village.", "Book in advance during school holidays."}),
			BestTime:     "Year-round (cave temperature is constant at 25°C)",
			Weather:      jm(map[string]string{"temp": "25°C", "condition": "Cave Interior", "status": "Constant comfortable temperature inside the cave."}),
			Latitude:     -8.0120,
			Longitude:    110.6100,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "Floating through the dark cave with glowing stalactites above was magical!"},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "pindul-p1", "name": "Pindul Adventure Tours", "category": "guide", "image": "https://images.unsplash.com/photo-1544717297-fa95b6ee9643?q=80&w=600", "rating": 4.8, "price": "IDR 100,000 / person", "distance": "On-Site", "description": "Professional cave tubing operators with 15+ years of experience.", "address": "Bejiharjo, Karangmojo, Gunungkidul"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Do I need to know how to swim?", "a": "No. Life jackets are provided and the water is calm."},
				{"q": "Is it safe?", "a": "Yes. Guides accompany every group and equipment is inspected daily."},
			}),
		},

		// ==========================================
		// 13. HUTAN PINUS MANGUNAN (NEW)
		// ==========================================
		{
			ExternalID:   "pinusmangunan",
			Name:         "Hutan Pinus Mangunan",
			Tagline:      "The Enchanted Pine Forest Above the Clouds",
			Category:     "nature",
			Location:     "Bantul, Yogyakarta",
			SubRegion:    "Bantul",
			Images:       ja([]string{"https://images.unsplash.com/photo-1654739595101-594699f889d4?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1470071459604-3b5ec3a7fe05?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.6,
			ReviewCount:  3100,
			Description:  "Hutan Pinus Mangunan is a vast pine forest perched on the hills of Mangunan, Bantul. Famous for its Instagram-worthy wooden platforms, swing seats, and the magical atmosphere created by tall pine trees stretching endlessly into the mist. On clear mornings, you can see the Indian Ocean from the viewpoints.",
			Story:        "The pine forest was originally a government reforestation project in the 1980s. Local residents later transformed it into an eco-tourism destination, building creative photo spots among the trees. It became a viral sensation on social media, attracting hundreds of thousands of visitors annually and boosting the local economy significantly.",
			TicketPrice:  "IDR 10,000 / person",
			OpeningHours: "06:00 AM – 05:00 PM Daily",
			Facilities:   ja([]string{"Pine Forest Walking Trails", "Instagram Photo Platforms", "Swing Seats", "Hilltop Viewpoint", "Cafes & Food Stalls", "Parking Area", "Restrooms", "Camping Area"}),
			TravelTips:   ja([]string{"Arrive before 07:00 AM for misty, ethereal photos.", "The hilltop viewpoint offers stunning sunrise views over the southern coast.", "Wear comfortable walking shoes for the forest trails.", "Combine with a visit to Becici Peak nearby."}),
			BestTime:     "Early morning (06:00–08:00 AM) for misty forest atmosphere",
			Weather:      jm(map[string]string{"temp": "25°C", "condition": "Misty Morning", "status": "Perfect conditions for atmospheric pine forest photos."}),
			Latitude:     -7.9520,
			Longitude:    110.3890,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Yuki Tanaka", "userAvatar": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150", "rating": 4.5, "date": "2026-07-02", "comment": "The misty morning photos from here are absolutely stunning. A photographer's paradise!"},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "pm-p1", "name": "Mangunan Pine Forest Cafe", "category": "cafe", "image": "https://images.unsplash.com/photo-1501339847302-ac426a4a7cbb?q=80&w=600", "rating": 4.5, "price": "IDR 15,000 – 35,000 / person", "distance": "On-Site", "description": "Rustic forest cafe serving traditional drinks and snacks.", "address": "Mangunan, Dlingo, Bantul"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is there an entrance fee?", "a": "Yes, IDR 10,000 per person. Parking IDR 5,000."},
				{"q": "Can I camp overnight?", "a": "Yes. Designated camping areas are available."},
			}),
		},

		// ==========================================
		// 14. BOROBUDUR TEMPLE (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "borobudur",
			Name:         "Borobudur Temple",
			Tagline:      "The World's Largest Buddhist Temple at Dawn",
			Category:     "heritage",
			Location:     "Magelang, Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1590211702270-5e64e0a2e04f?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1621869606578-1561708a7e09?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1592021369622-35e0f8d4491c?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.9,
			ReviewCount:  12500,
			Description:  "Borobudur is the world's largest Buddhist temple, built in the 9th century during the Sailendra dynasty. The monument consists of nine stacked platforms — six square and three circular — topped by a central dome. Over 2,672 relief panels and 504 Buddha statues adorn the temple, making it one of the greatest Buddhist monuments on earth. A UNESCO World Heritage Site since 1991.",
			Story:        "Borobudur lay hidden for centuries under volcanic ash and jungle growth after being abandoned in the 14th century. Sir Thomas Stamford Raffles rediscovered it in 1814 during the British occupation of Java. The Indonesian government and UNESCO led a massive restoration from 1975 to 1982, involving over 600 workers dismantling and reassembling the entire structure stone by stone.",
			TicketPrice:  "IDR 375,000 (Foreign) / IDR 50,000 (Domestic) / Sunrise package: IDR 450,000",
			OpeningHours: "06:30 AM – 05:00 PM Daily (Sunrise: 04:30 AM – 06:00 AM)",
			Facilities:   ja([]string{"Sunrise Viewing Platform", "Audio Guide Rental", "Buddhist Meditation Area", "Museum & Visitor Center", "Wheelchair Access", "Batik Workshops", "Souvenir Market", "Restaurants & Cafes", "Mosque & Prayer Room"}),
			TravelTips:   ja([]string{"Book the sunrise package to see the misty dawn over stupas.", "Wear a sarong (provided at entrance) out of respect for the sacred site.", "Hire a certified guide — relief panels tell the story of Buddhism.", "Arrive by 06:00 AM to avoid midday heat and crowds."}),
			BestTime:     "May to October (Dry Season) for sunrise; avoid weekends for fewer crowds",
			Weather:      jm(map[string]string{"temp": "25°C", "condition": "Clear Dawn", "status": "Excellent conditions for sunrise at Borobudur."}),
			Latitude:     -7.6079,
			Longitude:    110.2038,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-10", "comment": "Watching the sunrise over the misty stupas of Borobudur was the most spiritual experience of my life."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "bb-p1", "name": "Manohara Hotel Borobudur", "category": "hotel", "image": "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600", "rating": 4.6, "price": "IDR 800,000 / night", "distance": "0.5 km from temple", "description": "The only hotel within the Borobudur Archaeological Park — includes sunrise access.", "address": "Jl. Badrawarna, Borobudur, Magelang"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Can I climb to the top?", "a": "Yes. Visitors can climb to the top platform with a guide."},
				{"q": "Is the sunrise package worth it?", "a": "Absolutely — watching dawn break over the stupas is unforgettable."},
			}),
		},

		// ==========================================
		// 15. MENDUT TEMPLE (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "mendut",
			Name:         "Mendut Temple",
			Tagline:      "Home of the Majestic 3-Meter Seated Buddha",
			Category:     "heritage",
			Location:     "Magelang, Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1590211702270-5e64e0a2e04f?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1545569341-9eb8b30979d9?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1592021369622-35e0f8d4491c?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.7,
			ReviewCount:  1200,
			Description:  "Candi Mendut is a 9th-century Mahayana Buddhist temple located 3 km east of Borobudur. It houses a magnificent 3-meter seated Buddha statue in the bhumisparsha mudra (earth-touching pose), flanked by statues of the Bodhisattvas Avalokitesvara and Vajrapani — among the finest stone carvings of their era.",
			Story:        "Mendut is the oldest of the three temples that form the Borobudur archaeological axis — running east to west from Mendut to Pawon to Borobudur. This alignment symbolizes the journey from worldly desire (Mendut) to purification (Pawon) to enlightenment (Borobudur), and was used as the route for the annual Waisak procession.",
			TicketPrice:  "IDR 30,000 (Adult) / IDR 15,000 (Child)",
			OpeningHours: "07:00 AM – 05:00 PM Daily",
			Facilities:   ja([]string{"Buddha Statue Viewing Hall", "Meditation Area", "Small Museum", "Souvenir Stalls", "Parking Area", "Restrooms"}),
			TravelTips:   ja([]string{"Combine with Borobudur and Pawon in one morning.", "Visit during Waisak (May full moon) to see the torchlight procession.", "The interior Buddha statue is best photographed in the morning light.", "Part of the Borobudur ticket package — save your receipt."}),
			BestTime:     "May for Waisak festival; mornings year-round for best lighting",
			Weather:      jm(map[string]string{"temp": "27°C", "condition": "Partly Cloudy", "status": "Good conditions for temple exploration."}),
			Latitude:     -7.6049,
			Longitude:    110.2306,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Budi Santoso", "userAvatar": "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?q=80&w=150", "rating": 5, "date": "2026-06-12", "comment": "The giant Buddha inside is awe-inspiring — better than the photos suggest."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "mt-p1", "name": "Mendut Heritage Homestay", "category": "hotel", "image": "https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?q=80&w=600", "rating": 4.5, "price": "IDR 350,000 / night", "distance": "1 km", "description": "Traditional Javanese homestay steps from Mendut Temple.", "address": "Ds. Mendut, Mungkid, Magelang"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is Mendut included in the Borobudur ticket?", "a": "Yes, the Borobudur ticket package covers Mendut and Pawon."},
				{"q": "How long do I need to visit?", "a": "30-45 minutes is enough to appreciate the temple."},
			}),
		},

		// ==========================================
		// 16. PAWON TEMPLE (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "pawon",
			Name:         "Pawon Temple",
			Tagline:      "The Jewel of the Borobudur Sacred Axis",
			Category:     "heritage",
			Location:     "Magelang, Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1590211702270-5e64e0a2e04f?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1545569341-9eb8b30979d9?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1592021369622-35e0f8d4491c?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.6,
			ReviewCount:  850,
			Description:  "Candi Pawon is an elegant 9th-century Buddhist temple perfectly positioned between Mendut and Borobudur on the sacred east-west axis. Though smaller than its neighbors, Pawon is renowned for its exquisite proportions and intricate Kala-Makara carvings. The temple likely served as a storehouse for sacred objects.",
			Story:        "Pawon's name derives from the Javanese word 'pawon' (kitchen), perhaps referring to the food offerings once prepared here for the Waisak procession. The temple was rediscovered in 1896 and underwent restoration between 1903 and 1908. The reliefs on its walls depict kalpataru (tree of life) panels flanked by celestial beings.",
			TicketPrice:  "IDR 25,000 (Adult) / IDR 10,000 (Child)",
			OpeningHours: "07:00 AM – 04:00 PM Daily",
			Facilities:   ja([]string{"Temple Viewing Area", "Relief Panels", "Small Garden", "Parking Area", "Restrooms"}),
			TravelTips:   ja([]string{"A quick stop — 15-20 minutes is enough.", "Best light for photos is late afternoon.", "The kalpataru reliefs on the outer walls are exceptionally well-preserved.", "Often missed by tourists — you'll have it almost to yourself."}),
			BestTime:     "Afternoon for golden light; year-round",
			Weather:      jm(map[string]string{"temp": "28°C", "condition": "Sunny", "status": "Nice weather for a short temple visit."}),
			Latitude:     -7.6062,
			Longitude:    110.2192,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Yuki Tanaka", "userAvatar": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150", "rating": 4.5, "date": "2026-06-10", "comment": "Small but perfectly proportioned. The carvings are incredibly detailed."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "pw-p1", "name": "Pawon Art Village", "category": "attraction", "image": "https://images.unsplash.com/photo-1518895949257-7621c3c786d7?q=80&w=600", "rating": 4.3, "price": "Free entry", "distance": "0.3 km", "description": "Local artisan village with batik and wood carving workshops.", "address": "Ds. Wanurejo, Borobudur, Magelang"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Do I need a separate ticket?", "a": "No. It's covered by the Borobudur ticket package."},
				{"q": "Can I walk here from Borobudur?", "a": "Yes, it's a pleasant 2 km walk through rice paddies."},
			}),
		},

		// ==========================================
		// 17. PLAOSAN TEMPLE (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "plaosan",
			Name:         "Plaosan Temple",
			Tagline:      "The Twin Temple of Grace and Elegance",
			Category:     "heritage",
			Location:     "Klaten, Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1578469550956-0e16b69c6a3d?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.7,
			ReviewCount:  1100,
			Description:  "Candi Plaosan is a stunning 9th-century Buddhist temple complex located just 1 km east of Prambanan. Built during the reigns of Rakai Pikatan and Queen Pramodhawardhani of the Mataram Kingdom, the complex features twin main temples (Plaosan Lor and Plaosan Kidul), adorned with elegant reliefs and crowned with distinctive stupas.",
			Story:        "Plaosan was built as a symbol of reconciliation between the Hindu Sanjaya and Buddhist Sailendra dynasties, whose union was sealed through the marriage of King Pikatan (Hindu) to Princess Pramodhawardhani (Buddhist). The temple's unique design blends both traditions, with Hindu-style door guardians protecting a Buddhist sanctuary.",
			TicketPrice:  "IDR 30,000 (Adult) / IDR 15,000 (Child)",
			OpeningHours: "07:00 AM – 05:00 PM Daily",
			Facilities:   ja([]string{"Twin Temple Complex", "Buddhist Stupa Garden", "Relief Panel Walkways", "Parking Area", "Food Stalls", "Restrooms"}),
			TravelTips:   ja([]string{"Visit right after sunrise for golden light on the twin temples.", "Much quieter than Prambanan — you may have the place to yourself.", "The south temple (Plaosan Kidul) has the best-preserved reliefs.", "Combine with Prambanan and Sewu Temple in one trip."}),
			BestTime:     "May to October (Dry Season); early morning for best light and fewer crowds",
			Weather:      jm(map[string]string{"temp": "27°C", "condition": "Partly Cloudy", "status": "Good conditions for exploring the temple complex."}),
			Latitude:     -7.7467,
			Longitude:    110.5047,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Yuki Tanaka", "userAvatar": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150", "rating": 5, "date": "2026-06-18", "comment": "Less crowded than Prambanan but equally beautiful. The twin temples are stunning at sunrise."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "pl-p1", "name": "Plaosan Heritage Guide", "category": "guide", "image": "https://images.unsplash.com/photo-1544717297-fa95b6ee9643?q=80&w=600", "rating": 4.7, "price": "IDR 100,000 / tour", "distance": "On-Site", "description": "Local guides offering deep insight into Plaosan's unique Hindu-Buddhist fusion.", "address": "Plaosan, Prambanan, Klaten"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is Plaosan included in the Prambanan ticket?", "a": "No, it requires a separate ticket but is very affordable."},
				{"q": "How much time do I need?", "a": "About 1 hour to fully explore both temples and the grounds."},
			}),
		},

		// ==========================================
		// 18. SEWU TEMPLE (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "sewu",
			Name:         "Sewu Temple",
			Tagline:      "The Grand Buddhist Monastery of Prambanan",
			Category:     "heritage",
			Location:     "Klaten, Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1578469550956-0e16b69c6a3d?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1621869606578-1561708a7e09?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.7,
			ReviewCount:  1800,
			Description:  "Candi Sewu is an 8th-century Mahayana Buddhist temple complex just 800 meters north of Prambanan. The second-largest Buddhist temple compound in Indonesia after Borobudur, Sewu features a central temple surrounded by 240 smaller perwara (guardian) temples arranged in a mandala pattern, showcasing the pinnacle of Sailendra architecture.",
			Story:        "Sewu, meaning 'a thousand' in Javanese, refers to the countless guardian statues that once filled its 240 subsidiary temples. Built before Borobudur, Sewu was the largest Buddhist monastery complex in ancient Java. An inscription from 792 AD mentions King Indra of the Sailendra dynasty consecrating the temple to the bodhisattva Manjusri.",
			TicketPrice:  "IDR 30,000 (Adult) / IDR 15,000 (Child) — often covered by Prambanan combo ticket",
			OpeningHours: "07:00 AM – 05:00 PM Daily",
			Facilities:   ja([]string{"Main Temple (Central Shrine)", "240 Perwara Temples", "Buddha Statue Niches", "Stone Relief Corridors", "Parking Area", "Food Stalls", "Restrooms"}),
			TravelTips:   ja([]string{"Buy the Prambanan combo ticket that includes Sewu — great value.", "Early morning (07:00 AM) for the best photos with no crowds.", "Walk from Prambanan — it's only 800 meters through a pleasant path.", "Climb the central temple stairs for a view of the entire mandala layout."}),
			BestTime:     "May to October (Dry Season); morning for best lighting on the stone carvings",
			Weather:      jm(map[string]string{"temp": "28°C", "condition": "Sunny", "status": "Perfect for exploring the vast temple complex."}),
			Latitude:     -7.7439,
			Longitude:    110.4967,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "The mandala layout is breathtaking. I can't believe this gem is right next to Prambanan yet so quiet."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "sw-p1", "name": "Prambanan Heritage Hotel", "category": "hotel", "image": "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600", "rating": 4.4, "price": "IDR 700,000 / night", "distance": "2 km", "description": "Heritage hotel with direct views of Prambanan and Sewu temples.", "address": "Jl. Raya Solo-Yogya, Prambanan, Klaten"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is Sewu included in the Prambanan ticket?", "a": "Yes, if you buy the Prambanan + Sewu combo ticket."},
				{"q": "How long to explore Sewu?", "a": "Plan for 1–1.5 hours to fully appreciate the complex."},
			}),
		},

		// ==========================================
		// 19. ROWO JOMBOR (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "rowojombor",
			Name:         "Rowo Jombor",
			Tagline:      "Floating Stalls and Lakeside Serenity",
			Category:     "nature",
			Location:     "Klaten, Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1506905925346-21bda4d32df4?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1470071459604-3b5ec3a7fe05?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.4,
			ReviewCount:  1900,
			Description:  "Rowo Jombor is a scenic artificial lake in Klaten, famous for its unique floating warungs (food stalls) where visitors can enjoy fresh seafood and local specialties while seated on bamboo rafts floating on the water. Surrounded by lush green hills and rice paddies, it's a peaceful escape just 30 minutes from Yogyakarta.",
			Story:        "Rowo Jombor was originally a low-lying area that naturally flooded over time. Local fishermen began selling snacks from their boats, which evolved into the floating market concept. The floating warungs became a popular attraction, with visitors arriving by paddle boat to eat, shop, and enjoy the relaxed atmosphere on the water.",
			TicketPrice:  "IDR 10,000 / person (parking additional)",
			OpeningHours: "07:00 AM – 06:00 PM Daily",
			Facilities:   ja([]string{"Floating Warungs (Food Stalls)", "Paddle Boat Rentals", "Lakeside Walking Path", "Children's Play Area", "Parking Area", "Restrooms", "Mosque"}),
			TravelTips:   ja([]string{"Come hungry — the floating warungs serve excellent fresh fish, crabs, and fried tempeh.", "Rent a paddle boat to reach the floating stalls (IDR 30,000/hour).", "Late afternoon is the best time for cooler weather and sunset views.", "Weekdays are much quieter than weekends."}),
			BestTime:     "Late afternoon (03:00–05:00 PM) for sunset; weekdays for fewer crowds",
			Weather:      jm(map[string]string{"temp": "29°C", "condition": "Partly Cloudy", "status": "Good afternoon for boating and dining on the water."}),
			Latitude:     -7.7170,
			Longitude:    110.5830,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Budi Santoso", "userAvatar": "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?q=80&w=150", "rating": 4.5, "date": "2026-07-02", "comment": "Eating fresh grilled fish on a bamboo raft floating on the lake — such a unique experience!"},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "rj-p1", "name": "Rowo Jombor Floating Cafe", "category": "restaurant", "image": "https://images.unsplash.com/photo-1552566626-52f8b828add9?q=80&w=600", "rating": 4.3, "price": "IDR 30,000 – 70,000 / person", "distance": "On-Site", "description": "Popular floating warung serving fresh lake fish and traditional Javanese dishes.", "address": "Rowo Jombor, Klaten"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Can I swim in the lake?", "a": "Swimming is not recommended as it's used for fishing and boating."},
				{"q": "Are there vegetarian options at the floating warungs?", "a": "Yes, there are vegetarian options including fried tempeh and tahu."},
			}),
		},

		// ==========================================
		// 20. WADUK GAJAH MUNGKUR (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "gajahmungkur",
			Name:         "Waduk Gajah Mungkur",
			Tagline:      "The Great Reservoir of Wonogiri",
			Category:     "nature",
			Location:     "Wonogiri, Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1506905925346-21bda4d32df4?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1470071459604-3b5ec3a7fe05?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.5,
			ReviewCount:  2400,
			Description:  "Waduk Gajah Mungkur (Gajah Mungkur Reservoir) is a massive reservoir spanning over 8,800 hectares in Wonogiri. Built in the 1970s, it serves as a vital irrigation source and hydropower facility while offering stunning water landscapes with scattered islands. Visitors enjoy boat tours, water sports, camping, and bird watching against a backdrop of rolling hills.",
			Story:        "The reservoir was constructed between 1974 and 1981 as part of the Bengawan Solo river management project. The construction submerged several villages, and over 1,000 families were relocated. The name 'Gajah Mungkur' comes from a nearby hill that resembles a crouching elephant. Today, the reservoir has become a vital economic and tourism hub for Wonogiri.",
			TicketPrice:  "IDR 10,000 / person / IDR 50,000 – 200,000 for boat tours",
			OpeningHours: "06:00 AM – 06:00 PM Daily",
			Facilities:   ja([]string{"Reservoir Viewpoint", "Boat Tours (Motor & Paddle)", "Water Sports", "Camping Grounds", "Bird Watching Tower", "Lakeside Restaurants", "Parking Area", "Restrooms"}),
			TravelTips:   ja([]string{"Take the sunset boat tour for breathtaking views over the reservoir islands.", "Visit during the dry season when the water level is stable.", "Bring binoculars for bird watching — over 60 species have been recorded.", "The floating restaurants serve excellent fresh-water fish."}),
			BestTime:     "May to September (Dry Season) for boating; early morning for bird watching",
			Weather:      jm(map[string]string{"temp": "28°C", "condition": "Clear", "status": "Perfect conditions for a boat tour on the reservoir."}),
			Latitude:     -7.8970,
			Longitude:    110.9270,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Yuki Tanaka", "userAvatar": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150", "rating": 4.5, "date": "2026-06-22", "comment": "The vast reservoir with green islands scattered across the water is a beautiful sight. The sunset cruise was fantastic."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "gm-p1", "name": "Gajah Mungkur Boating Club", "category": "rental", "image": "https://images.unsplash.com/photo-1504280390367-361c6d9f38f4?q=80&w=600", "rating": 4.4, "price": "IDR 100,000 – 200,000 / tour", "distance": "On-Site", "description": "Professional boat tours with experienced local captains.", "address": "Jalan Waduk Gajah Mungkur, Wonogiri"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Can I camp at the reservoir?", "a": "Yes, designated camping areas are available by the lakeshore."},
				{"q": "Is it safe to swim?", "a": "Swimming is not recommended — it's a working reservoir."},
			}),
		},

		// ==========================================
		// 21. KERATON SURAKARTA (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "keratonsolo",
			Name:         "Keraton Surakarta",
			Tagline:      "The Royal Palace of Solo — Tradition Meets Modernity",
			Category:     "heritage",
			Location:     "Surakarta (Solo), Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1596402184320-417e7178b2cd?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1581456495146-65a71b2c8e52?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1566559631133-969041fc5583?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.6,
			ReviewCount:  3500,
			Description:  "Keraton Surakarta Hadiningrat is the official palace of the Susuhunan (King) of Surakarta, built in 1744 by Pakubuwono II. The palace complex showcases exquisite Javanese royal architecture, sacred heirlooms, gamelan instruments, and remains the cultural heart of Solo. The palace museum holds one of the finest collections of Javanese art in existence.",
			Story:        "The Keraton was built after the Mataram Kingdom split in 1755 (Giyanti Treaty), dividing the kingdom into Yogyakarta and Surakarta. The palace grounds include the main pendopo (audience hall), the royal meditation chamber, and the Sasana Sewaka hall. The palace still actively hosts traditional ceremonies and cultural performances.",
			TicketPrice:  "IDR 25,000 (Adult) / IDR 15,000 (Child) / Camera fee IDR 5,000",
			OpeningHours: "08:30 AM – 03:30 PM Tuesday–Sunday (Closed Monday)",
			Facilities:   ja([]string{"Royal Museum", "Gamelan Performance Hall", "Royal Carriage Collection", "Batik Museum", "Javanese Architecture Tour", "Gift Shop", "Clean Restrooms", "Parking Area"}),
			TravelTips:   ja([]string{"Visit on Saturday morning for the best cultural performances at the palace.", "Hire a guide — the stories behind the artifacts are fascinating.", "Don't miss the museum's keris (ceremonial dagger) collection.", "Combine with a walk through the nearby Klewer Market for batik shopping."}),
			BestTime:     "Tuesday to Sunday mornings for the full palace experience",
			Weather:      jm(map[string]string{"temp": "30°C", "condition": "Sunny", "status": "Warm day — perfect for exploring the palace and Solo's heritage streets."}),
			Latitude:     -7.5733,
			Longitude:    110.8267,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-07-01", "comment": "The palace is a living museum. The gamelan performance gave me chills — such a rich cultural experience."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "ks-p1", "name": "The Royal Surakarta Heritage Hotel", "category": "hotel", "image": "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600", "rating": 4.7, "price": "IDR 900,000 / night", "distance": "1.5 km", "description": "Luxury heritage hotel in the heart of Solo, walking distance from the Keraton.", "address": "Jl. Slamet Riyadi, Surakarta"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is it still an active palace?", "a": "Yes, the Susuhunan and his family reside in parts of the complex."},
				{"q": "What is the dress code?", "a": "Modest clothing is required. Sarongs are available at the entrance."},
			}),
		},

		// ==========================================
		// 22. BENTENG PENDEM (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "bentengpendem",
			Name:         "Benteng Pendem",
			Tagline:      "The Hidden Dutch Fortress on Turtle Bay",
			Category:     "heritage",
			Location:     "Cilacap, Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1584892077118-6399f5a200e5?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1566559631133-969041fc5583?do=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.4,
			ReviewCount:  900,
			Description:  "Benteng Pendem (The Sunken Fortress) is a Dutch colonial coastal fortress built between 1861 and 1879 on Teluk Penyu (Turtle Bay) in Cilacap. Designed to defend the strategic port from naval attacks, the fort is built low into the ground — hence its name — with thick coral-stone walls, underground tunnels, cannons, and a moat.",
			Story:        "Built by the Dutch East Indies government during the Aceh War period, Benteng Pendem was part of a coastal defense network. The fort appears 'sunken' because it was built at sea level with only the roof visible from the ocean, making it a difficult target for enemy ships. After Indonesian independence, it was used as a military barracks and later opened as a heritage site.",
			TicketPrice:  "IDR 15,000 (Adult) / IDR 10,000 (Child)",
			OpeningHours: "08:00 AM – 05:00 PM Daily",
			Facilities:   ja([]string{"Underground Tunnels", "Historical Cannon Displays", "Dutch Barracks", "Coastal Viewing Platform", "Small Museum", "Parking Area", "Food Stalls", "Restrooms"}),
			TravelTips:   ja([]string{"Wear a flashlight or use your phone — the underground tunnels are dark.", "Hire a guide at the entrance for the best historical context.", "Combine with a visit to Teluk Penyu beach right next door.", "The fort looks best in golden hour light (late afternoon)."}),
			BestTime:     "May to October (Dry Season); late afternoon for good light",
			Weather:      jm(map[string]string{"temp": "30°C", "condition": "Sunny with Sea Breeze", "status": "Good conditions for exploring the fort and nearby beach."}),
			Latitude:     -7.7410,
			Longitude:    109.0140,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Budi Santoso", "userAvatar": "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?q=80&w=150", "rating": 4.5, "date": "2026-07-15", "comment": "Exploring the underground tunnels and seeing the old Dutch cannons was fascinating. Great history lesson."},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "bp-p1", "name": "Cilacap Seaside Inn", "category": "hotel", "image": "https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?q=80&w=600", "rating": 4.2, "price": "IDR 350,000 / night", "distance": "2 km", "description": "Budget-friendly inn steps from Teluk Penyu beach.", "address": "Jl. Benteng Pendem, Cilacap"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "How long does the tour take?", "a": "About 1 hour for a full exploration of the fort and tunnels."},
				{"q": "Are the tunnels safe?", "a": "Yes, they are well-maintained with some lighting. Watch your step."},
			}),
		},

		// ==========================================
		// 23. TELUK PENYU (Near Yogyakarta)
		// ==========================================
		{
			ExternalID:   "telukpenyu",
			Name:         "Teluk Penyu Beach",
			Tagline:      "Turtle Bay — Where History Meets the Ocean",
			Category:     "nature",
			Location:     "Cilacap, Central Java",
			SubRegion:    "Near Yogyakarta",
			Images:       ja([]string{"https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1519046904884-53103b34b206?auto=format&fit=crop&w=1200&q=80", "https://images.unsplash.com/photo-1470071459604-3b5ec3a7fe05?auto=format&fit=crop&w=1200&q=80"}),
			Rating:       4.4,
			ReviewCount:  1700,
			Description:  "Teluk Penyu (Turtle Bay) is Cilacap's main beach, known for its golden sand, calm waters, and iconic views of Nusakambangan Island across the bay. The beach promenade is lined with food stalls, and the adjacent Benteng Pendem fortress adds a unique historical dimension to the coastal experience.",
			Story:        "Named after the sea turtles that once nested abundantly on its shores, Teluk Penyu has been a strategic anchorage since the Dutch colonial era. The bay is sheltered by Nusakambangan Island, creating calm waters perfect for recreation. The beach underwent major revitalization in 2015, transforming it into a modern public space while preserving its natural charm.",
			TicketPrice:  "IDR 10,000 / person",
			OpeningHours: "06:00 AM – 06:00 PM Daily",
			Facilities:   ja([]string{"Golden Sand Beach", "Beach Promenade", "Children's Playground", "Food Stalls & Warungs", "Benteng Pendem Adjacent", "Parking Area", "Restrooms", "Musholla"}),
			TravelTips:   ja([]string{"Visit late afternoon for sunset over Nusakambangan Island.", "Try the fresh seafood at the beachside warungs.", "Combine with Benteng Pendem for a full day of history and beach.", "The water is calm and safe for swimming."}),
			BestTime:     "May to September (Dry Season); late afternoon for sunset",
			Weather:      jm(map[string]string{"temp": "30°C", "condition": "Sunny with Sea Breeze", "status": "Beautiful beach weather — calm waters and sunny skies."}),
			Latitude:     -7.7430,
			Longitude:    109.0120,
			Reviews: ja([]map[string]interface{}{
				{"id": "r1", "userName": "Yuki Tanaka", "userAvatar": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150", "rating": 4.0, "date": "2026-06-25", "comment": "Clean beach with nice views of Nusakambangan. The seafood stalls are excellent!"},
			}),
			AffiliatePartners: ja([]map[string]interface{}{
				{"id": "tp-p1", "name": "Teluk Penyu Seafood Plaza", "category": "restaurant", "image": "https://images.unsplash.com/photo-1552566626-52f8b828add9?q=80&w=600", "rating": 4.3, "price": "IDR 30,000 – 80,000 / person", "distance": "On-Site", "description": "Beachfront seafood market with fresh catch grilled to order.", "address": "Teluk Penyu Beach, Cilacap"},
			}),
			FAQs: ja([]map[string]interface{}{
				{"q": "Is it safe to swim?", "a": "Yes, the bay is calm and protected by Nusakambangan Island."},
				{"q": "Are turtles still visible?", "a": "Sightings are rare now, but the beach is named after them."},
			}),
		},
	}
}
