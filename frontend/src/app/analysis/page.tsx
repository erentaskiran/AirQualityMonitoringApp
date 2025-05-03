"use client";

import { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

interface DensityData {
  [gridKey: string]: number; // e.g., "41.01_29.05": 5
}

export default function AnalysisPage() {
  const [minLat, setMinLat] = useState('');
  const [minLon, setMinLon] = useState('');
  const [maxLat, setMaxLat] = useState('');
  const [maxLon, setMaxLon] = useState('');
  const [densityData, setDensityData] = useState<DensityData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFetchDensity = async () => {
    if (!minLat || !minLon || !maxLat || !maxLon) {
      setError("Please fill in all bounding box coordinates.");
      return;
    }

    setLoading(true);
    setError(null);
    setDensityData(null);

    const params = new URLSearchParams({
      minLat,
      minLon,
      maxLat,
      maxLon,
    });

    try {
      // Assuming anomaly-processor runs on localhost:8000
      const response = await fetch(`http://localhost:8000/api/anomalies/density?${params.toString()}`);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: `HTTP error! status: ${response.status}` }));
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      const data: DensityData = await response.json();
      setDensityData(data);
      console.log("Fetched density data:", data);

    } catch (err: any) {
      console.error("Error fetching density data:", err);
      setError(err.message || "Failed to fetch density data.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1>Regional Anomaly Analysis</h1>
      <p>Enter bounding box coordinates to see anomaly density in that region.</p>

      <Card className="my-4">
        <CardHeader>
          <CardTitle>Define Region</CardTitle>
          <CardDescription>Enter the minimum and maximum latitude/longitude.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="minLat">Min Latitude</Label>
              <Input id="minLat" type="number" placeholder="e.g., 40.9" value={minLat} onChange={(e) => setMinLat(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="minLon">Min Longitude</Label>
              <Input id="minLon" type="number" placeholder="e.g., 28.9" value={minLon} onChange={(e) => setMinLon(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="maxLat">Max Latitude</Label>
              <Input id="maxLat" type="number" placeholder="e.g., 41.1" value={maxLat} onChange={(e) => setMaxLat(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="maxLon">Max Longitude</Label>
              <Input id="maxLon" type="number" placeholder="e.g., 29.1" value={maxLon} onChange={(e) => setMaxLon(e.target.value)} />
            </div>
          </div>
          <Button onClick={handleFetchDensity} disabled={loading}>
            {loading ? 'Loading...' : 'Analyze Density'}
          </Button>
          {error && <p className="text-red-500 text-sm mt-2">Error: {error}</p>}
        </CardContent>
      </Card>

      {densityData && (
        <Card>
          <CardHeader>
            <CardTitle>Anomaly Density Results</CardTitle>
            <CardDescription>Number of anomalies per grid cell (approx. coordinates).</CardDescription>
          </CardHeader>
          <CardContent>
            {Object.keys(densityData).length > 0 ? (
              <ul>
                {Object.entries(densityData).map(([gridKey, count]) => (
                  <li key={gridKey}>
                    Grid Cell ({gridKey.replace('_', ', ')}): {count} anomalies
                  </li>
                ))}
              </ul>
            ) : (
              <p>No anomalies found in the specified region.</p>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}