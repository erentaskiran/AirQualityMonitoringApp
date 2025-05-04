"use client";

import { useEffect, useState } from 'react';
import dynamic from 'next/dynamic';
import 'leaflet/dist/leaflet.css';
import { MakeRequest } from '@/lib/utils';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

const MapContainer = dynamic(() => import('react-leaflet').then((mod) => mod.MapContainer), { ssr: false });
const TileLayer = dynamic(() => import('react-leaflet').then((mod) => mod.TileLayer), { ssr: false });
const Marker = dynamic(() => import('react-leaflet').then((mod) => mod.Marker), { ssr: false });
const Popup = dynamic(() => import('react-leaflet').then((mod) => mod.Popup), { ssr: false });

const HeatmapLayer = dynamic(
  () => import('react-leaflet-heatmap-layer-v3').then((mod) => mod.HeatmapLayer), 
  { ssr: false }
);

interface Anomaly {
  latitude: number;
  longitude: number;
  value: number;
  parameter?: string;
  time?: string;
  description?: string;
}

interface DensityPoint {
  key: string;
  lat: number;
  lon: number;
  count: number;
}

export default function HeatmapPage() {
  const [anomalies, setAnomalies] = useState<Anomaly[]>([]);
  const [densityPoints, setDensityPoints] = useState<DensityPoint[]>([]);
  const [mapCenter, setMapCenter] = useState<[number, number]>([41.0, 29.0]); 
  const [isClient, setIsClient] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [mapBounds, setMapBounds] = useState({
    minLat: 40.8,
    minLon: 28.8,
    maxLat: 41.2,
    maxLon: 29.2
  });
  const [showHeatmap, setShowHeatmap] = useState(true);
  const [showMarkers, setShowMarkers] = useState(true);
  const [lastRefreshTime, setLastRefreshTime] = useState<Date | null>(null);
  const [shouldUpdate, setShouldUpdate] = useState(true);
  const [heatmapData, setHeatmapData] = useState<[number, number, number][]>([]);
  const [heatmapKey, setHeatmapKey] = useState<number>(0);

  const fetchDensityData = async () => {
    setLoading(true);
    setError(null);

    const params = new URLSearchParams({
      minLat: mapBounds.minLat.toString(),
      minLon: mapBounds.minLon.toString(),
      maxLat: mapBounds.maxLat.toString(),
      maxLon: mapBounds.maxLon.toString(),
    });

    try {
      const response = await fetch(`http://localhost:8081/api/anomalies/density?${params.toString()}`);
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data = await response.json();
      
      const points: DensityPoint[] = Object.entries(data).map(([key, count]) => {
        const [latStr, lonStr] = key.split('_');
        return {
          key,
          lat: parseFloat(latStr),
          lon: parseFloat(lonStr),
          count: count as number
        };
      });
      
      setDensityPoints(points);
      const newHeatmapData = points.map(point => [point.lat, point.lon, point.count * 10] as [number, number, number]);
      setHeatmapData(newHeatmapData);
      console.log("Fetched density data:", points);

    } catch (err) {
      console.error("Error fetching density data:", err);
      setError(`Failed to fetch density data: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  const fetchRecentAnomalies = async () => {
    try {
      const endTime = new Date().toISOString();
      const startTime = new Date(Date.now() - 12 * 60 * 60 * 1000).toISOString();
      
      console.log("Fetching anomalies...");
      const response = await fetch(`http://localhost:8081/api/anomalies/timerange`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "X-Start-Time": startTime,
          "X-End-Time": endTime
        }
      });
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data = await response.json();
      
      if (Array.isArray(data)) {
        setAnomalies(data);
        console.log(`Fetched ${data.length} anomalies`);
      } else {
        console.warn("Received non-array anomaly data:", data);
      }
    } catch (err) {
      console.error("Error fetching anomalies:", err);
    }
  };

  useEffect(() => {
    setIsClient(true);

    import('leaflet').then(L => {
      delete (L.Icon.Default.prototype as any)._getIconUrl;
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon-2x.png',
        iconUrl: 'https://unpkg.com/leaflet@1.7.1/dist/images/marker-icon.png',
        shadowUrl: 'https://unpkg.com/leaflet@1.7.1/dist/images/marker-shadow.png',
      });
    });

    navigator.geolocation.getCurrentPosition(
      (position) => {
        const userLat = position.coords.latitude;
        const userLon = position.coords.longitude;
        
        setMapCenter([userLat, userLon]);
        setMapBounds({
          minLat: userLat - 0.2,
          minLon: userLon - 0.2,
          maxLat: userLat + 0.2,
          maxLon: userLon + 0.2
        });

        console.log("User's location:", userLat, userLon);
      },
      (error) => {
        console.error("Error fetching location, using default:", error);
      }
    );

    fetchDensityData();
    fetchRecentAnomalies();
    setLastRefreshTime(new Date());
  }, []); 
  
  useEffect(() => {
    const debounceTimer = setTimeout(() => {
      if (isClient) { 
        fetchDensityData();
      }
    }, 1000); 
    
    return () => clearTimeout(debounceTimer);
  }, [mapBounds, isClient]);

  const handleRefresh = async () => {
    console.log("Manual refresh triggered");
    setLoading(true);
    
    try {
      setDensityPoints([]);
      setHeatmapData([]);
      setAnomalies([]);
      
      await fetchDensityData();
      await fetchRecentAnomalies();
      setLastRefreshTime(new Date());
      
      setHeatmapKey(prevKey => prevKey + 1);
    } catch (error) {
      console.error("Error during refresh:", error);
      setError(`Failed to refresh data: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  const toggleHeatmap = () => {
    setShowHeatmap(!showHeatmap);
  };

  const toggleMarkers = () => {
    setShowMarkers(!showMarkers);
  };

  if (!isClient) {
    return <div>Loading Map...</div>;
  }

  return (
    <div className="flex flex-col items-center w-full max-w-screen-xl mx-auto px-4">
      <h1 className="text-center my-4 text-2xl font-bold">Air Quality Heatmap</h1>
      <p className="text-center">Showing the intensity of recent air quality anomalies.</p>
      
      <div className="mb-4 flex gap-2">
        <Button onClick={handleRefresh} disabled={loading}>
          {loading ? 'Loading...' : 'Refresh Data'}
        </Button>
        <Button onClick={toggleHeatmap} variant="outline">
          {showHeatmap ? 'Hide Heatmap' : 'Show Heatmap'}
        </Button>
        <Button onClick={toggleMarkers} variant="outline">
          {showMarkers ? 'Hide Markers' : 'Show Markers'}
        </Button>
      </div>
      
      {error && (
        <Card className="mb-4 w-full">
          <CardContent className="pt-4 text-red-500">
            Error: {error}
          </CardContent>
        </Card>
      )}
      
      <div className="w-full">
        <MapContainer center={mapCenter} zoom={12} style={{ height: '70vh', width: '100%' }}>
          <TileLayer
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          />
          
          {showHeatmap && heatmapData.length > 0 && (
            <HeatmapLayer
              key={`heatmap-${heatmapKey}`}
              points={heatmapData}
              longitudeExtractor={(p: [number, number, number]) => p[1]}
              latitudeExtractor={(p: [number, number, number]) => p[0]}
              intensityExtractor={(p: [number, number, number]) => p[2]}
              radius={25} 
              blur={15}
              max={50}
            />
          )}
          
          {showMarkers && anomalies.map((anomaly, idx) => (
            anomaly.latitude && anomaly.longitude ? (
              <Marker 
                key={`anomaly-${idx}`}
                position={[anomaly.latitude, anomaly.longitude]}
              >
                <Popup>
                  <div>
                    <h3 className="font-bold">{anomaly.parameter}</h3>
                    <p>Value: {anomaly.value}</p>
                    <p>Reason: {anomaly.description}</p>
                    {anomaly.time && <p>Time: {new Date(anomaly.time).toLocaleString()}</p>}
                  </div>
                </Popup>
              </Marker>
            ) : null
          ))}
          
          <Marker position={mapCenter}>
            <Popup>Your Location</Popup>
          </Marker>
        </MapContainer>
      </div>
      
      <div className="mt-4 text-center w-full">
        <p className="text-sm">
          Showing {densityPoints.length} density points and {anomalies.length} individual anomalies.
        </p>
        {loading && <p className="text-sm text-blue-500">Updating data...</p>}
        {lastRefreshTime && (
          <p className="text-sm text-gray-500">
            Last updated: {lastRefreshTime.toLocaleTimeString()}
          </p>
        )}
      </div>
    </div>
  );
}