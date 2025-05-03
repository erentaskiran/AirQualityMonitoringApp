"use client";

import { useEffect, useState } from 'react';
import dynamic from 'next/dynamic';
import 'leaflet/dist/leaflet.css';
// import 'leaflet.heat'; // Import leaflet.heat - Commented out as heatmap layer is removed

// Dynamically import Leaflet components
const MapContainer = dynamic(() => import('react-leaflet').then((mod) => mod.MapContainer), { ssr: false });
const TileLayer = dynamic(() => import('react-leaflet').then((mod) => mod.TileLayer), { ssr: false });
// Use a type assertion for the heatmap layer import if necessary - Removed heatmap layer import
// const HeatmapLayer = dynamic(() => import('react-leaflet-heatmap-layer-v3').then((mod) => mod.HeatmapLayer), {
//   ssr: false,
// }) as any; // Using 'as any' temporarily if type issues arise with v3 beta

interface Anomaly {
  latitude: number;
  longitude: number;
  value: number; // Intensity value for the heatmap
  parameter?: string;
  time?: string;
  description?: string;
}

export default function HeatmapPage() {
  const [anomalies, setAnomalies] = useState<Anomaly[]>([]);
  const [mapCenter, setMapCenter] = useState<[number, number]>([41.0, 29.0]); // Default center (Istanbul)
  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    setIsClient(true); // Indicate that we are on the client side

    // Fix Leaflet default icon issue (important for Marker if used later)
    import('leaflet').then(L => {
      delete (L.Icon.Default.prototype as any)._getIconUrl;
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon-2x.png',
        iconUrl: 'https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon.png',
        shadowUrl: 'https://unpkg.com/leaflet@1.7.1/dist/images/marker-shadow.png',
      });
    });

    // Get user's location to center the map (optional, fallback to default)
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setMapCenter([position.coords.latitude, position.coords.longitude]);
        console.log("User's location:", position.coords.latitude, position.coords.longitude);
      },
      (error) => {
        console.error("Error fetching location, using default:", error);
        // Keep default center if geolocation fails
      }
    );

    // WebSocket for historical/recent anomalies
    const historySocket = new WebSocket("ws://localhost:8000/ws/anomalys");

    historySocket.onopen = () => {
        console.log("WebSocket connection (anomalies) opened");
    };

    historySocket.onmessage = (event) => {
      try {
        const receivedData = JSON.parse(event.data);
        if (Array.isArray(receivedData)) {
          // Ensure data has lat, lon, and value for heatmap
          const validAnomalies = receivedData.filter(
            (a: any): a is Anomaly => a.latitude && a.longitude && typeof a.value === 'number'
          );
          setAnomalies(validAnomalies);
          // console.log("Received anomalies for heatmap:", validAnomalies); // Log might be less relevant now
        } else {
             console.warn("Received non-array data from WebSocket:", receivedData);
        }
      } catch (error) {
        console.error("Error parsing anomaly data:", error);
      }
    };

    historySocket.onerror = (error) => {
      console.error("WebSocket error (anomalies):", error);
    };

    historySocket.onclose = () => {
      console.log("WebSocket connection (anomalies) closed");
    };

    // Cleanup function
    return () => {
      historySocket.close();
    };
  }, []); // Run only once on component mount

  // Prepare data for HeatmapLayer: array of [lat, lng, intensity] - Removed heatmap data preparation
  // const heatmapData: [number, number, number][] = anomalies.map(a => [a.latitude, a.longitude, a.value]);

  if (!isClient) {
    // Render placeholder or loading state on the server
    return <div>Loading Map...</div>;
  }

  return (
    <div>
      <h1>Air Quality Heatmap</h1>
      <p>Showing the intensity of recent air quality anomalies. (Heatmap temporarily disabled)</p> {/* Updated text */}
      <MapContainer center={mapCenter} zoom={10} style={{ height: '70vh', width: '100%' }}>
        <TileLayer
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        />
        {/* Removed HeatmapLayer component */}
        {/* {heatmapData.length > 0 && (
          <HeatmapLayer
            points={heatmapData}
            longitudeExtractor={(p: [number, number, number]) => p[1]}
            latitudeExtractor={(p: [number, number, number]) => p[0]}
            intensityExtractor={(p: [number, number, number]) => p[2]}
            // Adjust these options as needed
            radius={20} // Radius of influence for each point
            blur={15}   // Blur effect
            max={50}    // Maximum intensity value (adjust based on your data range)
          />
        )} */}
      </MapContainer>
    </div>
  );
}