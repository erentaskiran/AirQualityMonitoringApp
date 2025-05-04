import { useEffect, useRef } from 'react';
import { useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet.heat';

// Import the type from our declaration file
import type { HeatmapLayerProps } from 'react-leaflet-heatmap-layer-v3';

// Custom React 19 compatible HeatmapLayer component
export function HeatmapLayer({
  points,
  longitudeExtractor,
  latitudeExtractor,
  intensityExtractor = () => 1,
  radius = 30,
  max = 3.0,
  minOpacity = 0.05,
  blur = 15,
  gradient = { 0.4: 'blue', 0.6: 'cyan', 0.7: 'lime', 0.8: 'yellow', 1.0: 'red' }
}: HeatmapLayerProps) {
  const map = useMap();
  const heatLayerRef = useRef<any>(null);
  
  useEffect(() => {
    // Clean up existing layer if it exists
    if (heatLayerRef.current) {
      heatLayerRef.current.remove();
    }

    // Transform points to the format expected by leaflet.heat
    const heatPoints = points.map(p => {
      const lat = latitudeExtractor(p);
      const lng = longitudeExtractor(p);
      const intensity = intensityExtractor(p);
      return [lat, lng, intensity];
    });

    // Create and add the heat layer to the map
    const heatLayer = (L as any).heatLayer(heatPoints, {
      radius,
      max,
      minOpacity,
      blur,
      gradient
    }).addTo(map);

    // Store reference for cleanup
    heatLayerRef.current = heatLayer;

    // Cleanup on component unmount
    return () => {
      if (heatLayerRef.current) {
        heatLayerRef.current.remove();
      }
    };
  }, [
    points, 
    longitudeExtractor, 
    latitudeExtractor, 
    intensityExtractor,
    radius,
    max,
    minOpacity,
    blur,
    gradient,
    map
  ]);

  // This component doesn't render anything directly
  return null;
}

export default HeatmapLayer;