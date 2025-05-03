"use client";

import { useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge"; // Import Badge if you want to use it

interface AnomalyAlert {
  parameter: string;
  value: number;
  time: string; // Keep as string, format for display
  latitude?: number;
  longitude?: number;
  description: string; // Reason for the anomaly
}

export default function AlertsPage() {
  const [alerts, setAlerts] = useState<AnomalyAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const ws = new WebSocket("ws://localhost:8000/ws/anomalys");
    setLoading(true);
    setError(null);

    ws.onopen = () => {
      console.log("WebSocket connection (alerts) opened");
    };

    ws.onmessage = (event) => {
      try {
        const receivedData: any[] = JSON.parse(event.data); // Expecting array from /ws/anomalys
        if (Array.isArray(receivedData)) {
          // Validate and map data
          const validAlerts = receivedData
            .map((item): AnomalyAlert | null => {
              // Basic validation
              if (!item.parameter || typeof item.value !== 'number' || !item.time || !item.description) {
                console.warn("Skipping invalid alert item:", item);
                return null;
              }
              return {
                parameter: item.parameter,
                value: item.value,
                time: item.time,
                latitude: item.latitude,
                longitude: item.longitude,
                description: item.description,
              };
            })
            .filter((item): item is AnomalyAlert => item !== null) // Filter out nulls
            .sort((a, b) => new Date(b.time).getTime() - new Date(a.time).getTime()); // Sort newest first

          setAlerts(validAlerts);
          console.log("Received alerts:", validAlerts);
        } else {
             console.warn("Received non-array data from WebSocket:", receivedData);
             setError("Received unexpected data format.");
        }
        setLoading(false);
      } catch (err) {
        console.error("Error processing WebSocket data:", err);
        setError("Failed to process alert data.");
        setLoading(false);
      }
    };

    ws.onerror = (err) => {
      console.error("WebSocket error (alerts):", err);
      setError("WebSocket connection error.");
      setLoading(false);
    };

    ws.onclose = () => {
      console.log("WebSocket connection (alerts) closed");
       if (loading) {
           setError("WebSocket connection closed before receiving data.");
           setLoading(false);
       }
    };

    // Cleanup function
    return () => {
      ws.close();
    };
  }, []); // Run only once

  const formatTime = (timeString: string) => {
    try {
      return new Date(timeString).toLocaleString();
    } catch (e) {
      return timeString; // Fallback to original string if parsing fails
    }
  };

  return (
    <div>
      <h1>Anomaly Alerts</h1>
      <p>Showing recent air quality anomalies detected by the system.</p>

      {loading && <p>Loading alerts...</p>}
      {error && <p className="text-red-500">Error: {error}</p>}

      {!loading && !error && alerts.length === 0 && (
        <p>No recent alerts found.</p>
      )}

      {!loading && !error && alerts.length > 0 && (
        <div className="space-y-4 mt-4">
          {alerts.map((alert, index) => (
            <Card key={index}>
              <CardHeader>
                <CardTitle>Anomaly Detected: {alert.parameter}</CardTitle>
                <CardDescription>
                  Reason: {alert.description} | Time: {formatTime(alert.time)}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p>Value: {alert.value.toFixed(2)}</p>
                {alert.latitude && alert.longitude && (
                  <p>Location: Lat {alert.latitude.toFixed(4)}, Lon {alert.longitude.toFixed(4)}</p>
                )}
                {/* Optional: Add a badge for the reason */}
                {/* <Badge variant="destructive">{alert.description}</Badge> */}
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}