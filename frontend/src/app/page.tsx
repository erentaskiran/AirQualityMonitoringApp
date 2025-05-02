"use client";
import { useEffect, useState } from 'react';

export default function Home() {
  const [anomalies, setAnomalies] = useState<any[]>([]);
  const [location, setLocation] = useState<{ latitude: number; longitude: number } | null>(null);

  useEffect(() => {
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
      }
    );
  }, []);

  useEffect(() => {
    if (!location) return;

    const socket = new WebSocket("ws://localhost:8000/ws/anomalys");

    socket.onmessage = (event) => {
      const receivedAnomalies = JSON.parse(event.data);
      setAnomalies(receivedAnomalies);
    };

    socket.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    socket.onclose = () => {
      console.log("WebSocket connection closed");
    };

    return () => {
      socket.close();
    };
  }, [location]);

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
        <ul>
          {anomalies.map((anomaly, index) => (
            <li key={index}>
              <strong>{anomaly.parameter}</strong>: {anomaly.value} at {anomaly.time}
              <br />
              Location: ({anomaly.latitude}, {anomaly.longitude})
              <br />
              Description: {anomaly.description}
            </li>
          ))}
        </ul>
      ) : (
        <p>Fetching your location...</p>
      )}
    </div>
  );
}
