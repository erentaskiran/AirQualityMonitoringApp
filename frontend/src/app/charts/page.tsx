"use client";

import { useEffect, useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { MakeRequest } from '@/lib/utils';

interface AnomalyData {
  time: string; // Keep as string initially
  value: number;
  parameter: string;
  // Add other fields if needed
}

interface ChartDataPoint {
  timestamp: number; // Store time as epoch milliseconds for sorting/charting
  timeLabel: string; // Formatted time string for display
  value: number;
  parameter: string;
}

export default function ChartsPage() {
  const [chartData, setChartData] = useState<ChartDataPoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    
    // Calculate timerange - last 24 hours
    const endTime = new Date().toISOString();
    const startTime = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString();
    
    // Fetch anomalies from API instead of WebSocket
    const fetchAnomalies = async () => {
      try {
        const response = await MakeRequest(
          'api/anomalies/timerange', 
          "GET", 
          null, 
          {
            "X-Start-Time": startTime,
            "X-End-Time": endTime
          }
        );
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const receivedData = await response.json();
        
        if (Array.isArray(receivedData)) {
          const processedData = receivedData
            .map(item => {
              const date = new Date(item.time);
              return {
                timestamp: date.getTime(), // For sorting and potential XAxis type='number'
                timeLabel: date.toLocaleString(), // For Tooltip/XAxis display
                value: item.value,
                parameter: item.parameter,
              };
            })
            .filter(item => !isNaN(item.timestamp) && typeof item.value === 'number') // Ensure valid data
            .sort((a, b) => a.timestamp - b.timestamp); // Sort by time ascending

          setChartData(processedData);
          console.log("Processed chart data:", processedData);
        } else {
          console.warn("Received non-array data from API:", receivedData);
          setError("Received unexpected data format.");
        }
      } catch (err) {
        console.error("Error fetching anomalies:", err);
        setError(`Failed to fetch anomaly data: ${err instanceof Error ? err.message : 'Unknown error'}`);
      } finally {
        setLoading(false);
      }
    };
    
    fetchAnomalies();
    
    // Optional: Set up polling to refresh data
    const intervalId = setInterval(fetchAnomalies, 30000); // Refresh every 30 seconds
    
    return () => {
      clearInterval(intervalId);
    };
  }, []); // Run only once

  return (
    <div>
      <h1>Air Quality Charts</h1>
      <p>Showing recent air quality anomaly values over time.</p>

      {loading && <p>Loading chart data...</p>}
      {error && <p style={{ color: 'red' }}>Error: {error}</p>}

      {!loading && !error && chartData.length === 0 && (
        <p>No anomaly data available to display.</p>
      )}

      {!loading && !error && chartData.length > 0 && (
        <div style={{ width: '100%', height: 400 }}> {/* Container for responsiveness */}
          <ResponsiveContainer>
            <LineChart
              data={chartData}
              margin={{
                top: 5,
                right: 30,
                left: 20,
                bottom: 5,
              }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              {/* Using timestamp for XAxis dataKey allows numerical sorting/scaling */}
              {/* Formatting the tick labels using timeLabel */}
              <XAxis
                 dataKey="timestamp"
                 type="number" // Treat timestamp as a number
                 scale="time" // Use time scale
                 domain={['dataMin', 'dataMax']} // Auto-adjust domain
                 tickFormatter={(unixTime) => new Date(unixTime).toLocaleTimeString()} // Format ticks
                 name="Time"
               />
              <YAxis label={{ value: 'Value', angle: -90, position: 'insideLeft' }} />
              <Tooltip labelFormatter={(label) => new Date(label).toLocaleString()} />
              <Legend />
              {/* Plotting all anomaly values on one line for now */}
              {/* Consider grouping by parameter later */}
              <Line
                 type="monotone"
                 dataKey="value"
                 stroke="#8884d8"
                 activeDot={{ r: 8 }}
                 name="Anomaly Value"
                 connectNulls // Connect points across gaps if needed
               />
               {/* Example for adding another parameter line later:
               <Line type="monotone" dataKey="pm25Value" stroke="#82ca9d" name="PM2.5" />
               */}
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}