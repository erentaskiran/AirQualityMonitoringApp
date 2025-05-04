"use client";
import dynamic from 'next/dynamic';
import { useEffect, useState } from 'react';

// Dynamically import react-leaflet components to prevent SSR issues
const MapContainer = dynamic(() => import('react-leaflet').then((mod) => mod.MapContainer), { ssr: false });
const TileLayer = dynamic(() => import('react-leaflet').then((mod) => mod.TileLayer), { ssr: false });
const Marker = dynamic(() => import('react-leaflet').then((mod) => mod.Marker), { ssr: false });
const Popup = dynamic(() => import('react-leaflet').then((mod) => mod.Popup), { ssr: false });

import 'leaflet/dist/leaflet.css';

export default function Home() {
  const [anomalies, setAnomalies] = useState<any[]>([]);
  const [location, setLocation] = useState<{ latitude: number; longitude: number } | null>(null);

  useEffect(() => {
    import('leaflet').then(L => {
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
        setLocation({ latitude: 41.0, longitude: 29.0 }); // Default to Istanbul
      }
    );
  }, []);

  useEffect(() => {
    if (!location) return;

    const liveSocket = new WebSocket("ws://localhost:8080/ws/live");

    liveSocket.onmessage = (event) => {
      try {
        const receivedData = JSON.parse(event.data);
        console.log("Received live data:", receivedData);
        
        const dataItems = Array.isArray(receivedData) ? receivedData : [receivedData];
        
        dataItems.forEach(item => {
          if (item && typeof item === 'object' && item.latitude && item.longitude) {
            const distance = calculateDistance(
              location.latitude,
              location.longitude,
              item.latitude,
              item.longitude
            );
            if (distance <= 25) {
              setAnomalies((prevAnomalies) => [...prevAnomalies, item]);
            }
          }
        });
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
    <div className="flex flex-col items-center w-full max-w-screen-xl mx-auto px-4">
      <h1 className="text-center my-4 text-2xl font-bold">Nearby Anomalies</h1>
      {location ? (
        <div className="w-full">
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
                    Time: {new Date(anomaly.time).toLocaleDateString()} {new Date(anomaly.time).toLocaleTimeString()}
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
        </div>
      ) : (
        <p className="text-center my-4">Fetching your location or using default...</p>
      )}
    </div>
  );
}
