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

    // WebSocket for historical anomalies
    const historySocket = new WebSocket("ws://localhost:8000/ws/anomalys");

    historySocket.onmessage = (event) => {
      const receivedData = JSON.parse(event.data);

      if (Array.isArray(receivedData)) {
        // Initial anomalies from the last 2 hours
        setAnomalies(receivedData);
      }
    };

    historySocket.onerror = (error) => {
      console.error("WebSocket error (history):", error);
    };

    historySocket.onclose = () => {
      console.log("WebSocket connection (history) closed");
    };

    // WebSocket for live anomalies
    const liveSocket = new WebSocket("ws://localhost:8000/ws/live");

    liveSocket.onmessage = (event) => {
      const receivedData = JSON.parse(event.data);

      // Live anomaly data
      setAnomalies((prevAnomalies) => [...prevAnomalies, receivedData]);
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
