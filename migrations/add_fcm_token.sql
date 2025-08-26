-- Migration: Add FCM Token column to drivers table
-- Date: 2024-12-19
-- Description: Add FCM token column for push notifications

ALTER TABLE drivers ADD COLUMN fcm_token VARCHAR(255);

-- Add index for better performance
CREATE INDEX idx_drivers_fcm_token ON drivers(fcm_token);

-- Add comment to column
COMMENT ON COLUMN drivers.fcm_token IS 'Firebase Cloud Messaging token for push notifications';
