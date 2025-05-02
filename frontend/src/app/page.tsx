"use client";
import dynamic from 'next/dynamic';
import { useEffect, useState } from 'react';

// Dynamically import react-leaflet components to prevent SSR issues
const MapContainer = dynamic(() => import('react-leaflet').then((mod) => mod.MapContainer), { ssr: false });
const TileLayer = dynamic(() => import('react-leaflet').then((mod) => mod.TileLayer), { ssr: false });
const Marker = dynamic(() => import('react-leaflet').then((mod) => mod.Marker), { ssr: false });
const Popup = dynamic(() => import('react-leaflet').then((mod) => mod.Popup), { ssr: false });

// Import Leaflet CSS
import 'leaflet/dist/leaflet.css';
// NOTE: Do not import L directly here to avoid SSR issues

export default function Home() {
  const [anomalies, setAnomalies] = useState<any[]>([]);
  const [location, setLocation] = useState<{ latitude: number; longitude: number } | null>(null);

  useEffect(() => {
    // This code runs only on the client
    import('leaflet').then(L => {
      // Fix for default icon issue with Webpack
      delete (L.Icon.Default.prototype as any)._getIconUrl;

      L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon-2x.png',
        iconUrl: 'https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon.png',
        shadowUrl: 'https://unpkg.com/leaflet@1.7.1/dist/images/marker-shadow.png',
      });
    });

    // Get user's current location
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setLocation({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
        });
        console.log("User's location:", position.coords.latitude, position.coords.longitude);
      },
      (error) => {
        console.error("Error fetching location:", error);
        // Fallback location if geolocation fails or is denied
        setLocation({ latitude: 41.0, longitude: 29.0 }); // Default to Istanbul
      }
    );
  }, []);

  useEffect(() => {
    if (!location) return;

    // WebSocket for historical anomalies
    const historySocket = new WebSocket("ws://localhost:8000/ws/anomalys");

    historySocket.onmessage = (event) => {
      try {
        const receivedData = JSON.parse(event.data);
        if (Array.isArray(receivedData)) {
          // Filter anomalies within 25km radius
          const filteredAnomalies = receivedData.filter((anomaly: any) => {
            if (anomaly.latitude && anomaly.longitude) {
              const distance = calculateDistance(
                location.latitude,
                location.longitude,
                anomaly.latitude,
                anomaly.longitude
              );
              return distance <= 25; // 25km radius
            }
            return false;
          });
          setAnomalies(filteredAnomalies);
        }
      } catch (error) {
        console.error("Error parsing history data:", error);
      }
    };

    historySocket.onerror = (error) => {
      console.error("WebSocket error (history):", error);
    };

    historySocket.onclose = () => {
      console.log("WebSocket connection (history) closed");
    };

    // WebSocket for live anomalies
    const liveSocket = new WebSocket("ws://localhost:8000/ws/live"); // Assuming /ws sends live data

    liveSocket.onmessage = (event) => {
      try {
        const receivedData = JSON.parse(event.data);
        // Check if it's a single anomaly object and has location
        if (receivedData && typeof receivedData === 'object' && !Array.isArray(receivedData) && receivedData.latitude && receivedData.longitude) {
          // Filter live anomaly within 25km radius
          const distance = calculateDistance(
            location.latitude,
            location.longitude,
            receivedData.latitude,
            receivedData.longitude
          );
          if (distance <= 25) {
            setAnomalies((prevAnomalies) => [...prevAnomalies, receivedData]);
          }
        }
      } catch (error) {
        console.error("Error parsing live data:", error);
      }
    };

    liveSocket.onerror = (error) => {
      console.error("WebSocket error (live):", error);
    };

    liveSocket.onclose = () => {
      console.log("WebSocket connection (live) closed");
    };

    return () => {
      historySocket.close();
      liveSocket.close();
    };
  }, [location]);

  // Haversine formula to calculate distance between two points
  const calculateDistance = (lat1: number, lon1: number, lat2: number, lon2: number) => {
    const toRad = (value: number) => (value * Math.PI) / 180;
    const R = 6371; // Earth's radius in km
    const dLat = toRad(lat2 - lat1);
    const dLon = toRad(lon2 - lon1);
    const a =
      Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) *
      Math.sin(dLon / 2) * Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c; // Distance in km
  };

  return (
    <div>
      <h1>Nearby Anomalies</h1>
      {location ? (
        <MapContainer center={[location.latitude, location.longitude]} zoom={11} style={{ height: "600px", width: "100%" }}>
          <TileLayer
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          />
          {anomalies.map((anomaly, index) => (
            // Ensure anomaly has valid latitude and longitude before rendering Marker
            anomaly.latitude && anomaly.longitude ? (
              <Marker key={index} position={[anomaly.latitude, anomaly.longitude]}>
                <Popup>
                  <strong>{anomaly.parameter}</strong>: {anomaly.value}
                  <br />
                  Time: {new Date(anomaly.time).toLocaleString()}
                  <br />
                  Description: {anomaly.description}
                </Popup>
              </Marker>
            ) : null
          ))}
           {/* Marker for user's location */}
           <Marker position={[location.latitude, location.longitude]}>
             <Popup>Your Location</Popup>
           </Marker>
        </MapContainer>
      ) : (
        <p>Fetching your location or using default...</p>
      )}
    </div>
  );
}
