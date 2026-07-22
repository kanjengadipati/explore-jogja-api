-- Clear broken Wikimedia Special:FilePath image URLs that return 404
UPDATE events SET image_url = NULL WHERE image_url LIKE '%commons.wikimedia.org/wiki/Special:FilePath/Jogja_Rockarta_Festival_2023.jpg';
UPDATE events SET image_url = NULL WHERE image_url LIKE '%commons.wikimedia.org/wiki/Special:FilePath/Keroncong_Plesiran_Mangunan.jpg';
