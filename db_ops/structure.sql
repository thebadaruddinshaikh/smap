use potholes;

-- Drop trigger if it exists
DROP TRIGGER IF EXISTS before_location_insert;

-- Create pothole_locations table if it doesn't exist
CREATE TABLE IF NOT EXISTS pothole_locations (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL COMMENT 'Location name or description',
    latitude DECIMAL(10,8) NOT NULL COMMENT 'Latitude in decimal degrees',
    longitude DECIMAL(11,8) NOT NULL COMMENT 'Longitude in decimal degrees',
    coordinates POINT NOT NULL SRID 4326 COMMENT 'Spatial point data',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add spatial index if it doesn't exist
SELECT IF(
    (
        -- Check if index exists
        SELECT COUNT(*)
        FROM information_schema.statistics
        WHERE table_schema = DATABASE()
        AND table_name = 'pothole_locations'
        AND index_name = 'idx_coordinates'
    ) = 0,
    'CREATE SPATIAL INDEX idx_coordinates ON pothole_locations (coordinates)',
    'SELECT 1'
) INTO @sql;
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Create trigger to automatically set the POINT data
DELIMITER //

CREATE TRIGGER IF NOT EXISTS before_location_insert 
BEFORE INSERT ON pothole_locations
FOR EACH ROW
BEGIN
    -- SRID 4326 specifically refers to the WGS 84 coordinate system - this is what GPS uses
    SET NEW.coordinates = ST_SRID(POINT(NEW.longitude, NEW.latitude), 4326);
END;//

DELIMITER ;
