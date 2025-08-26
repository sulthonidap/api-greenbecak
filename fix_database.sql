-- Fix database by dropping problematic table
DROP TABLE IF EXISTS driver_locations;

-- Recreate the table with correct structure
CREATE TABLE driver_locations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    driver_id INT NOT NULL,
    latitude DOUBLE NOT NULL,
    longitude DOUBLE NOT NULL,
    accuracy DOUBLE,
    speed DOUBLE,
    heading DOUBLE,
    is_online BOOLEAN DEFAULT FALSE,
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_driver_locations_driver_id (driver_id),
    INDEX idx_driver_locations_online (is_online),
    INDEX idx_driver_locations_last_seen (last_seen),
    INDEX idx_driver_locations_deleted_at (deleted_at),
    FOREIGN KEY (driver_id) REFERENCES drivers(id)
);
