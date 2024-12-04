package db_ops

const CHECK_POTHOLE_EXISTS = `
SELECT COUNT(*) as total_potholes
FROM pothole_locations
WHERE ST_Distance_Sphere(
    coordinates,
    ST_GeomFromText(CONCAT('POINT(', ?, ' ', ?, ')'), 4326)
) <= ?;
`

const INSERT_POTHOLE_QUERY = "INSERT INTO pothole_locations (name, latitude, longitude) VALUES (?, ?, ?);"

const GET_ALL_POTHOLE_QUERY = "SELECT * FROM pothole_locations"

const GET_VEHICLE_POTHOLES_QUERY = `
SELECT 
    latitude,
    longitude,
    ST_Distance_Sphere(
        coordinates,
        ST_GeomFromText(CONCAT('POINT(', ?, ' ', ?, ')'), 4326)
    ) as distance_m
FROM pothole_locations
WHERE ST_Distance_Sphere(
    coordinates,
    ST_GeomFromText(CONCAT('POINT(', ?, ' ', ?, ')'), 4326)
) <= ?
ORDER BY distance_m ASC;
`
