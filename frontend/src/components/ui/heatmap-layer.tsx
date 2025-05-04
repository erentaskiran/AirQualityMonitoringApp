import { useEffect, useRef } from 'react';
import { useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet.heat';

import type { HeatmapLayerProps } from 'react-leaflet-heatmap-layer-v3';

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
    if (heatLayerRef.current) {
      heatLayerRef.current.remove();
    }

    const heatPoints = points.map(p => {
      const lat = latitudeExtractor(p);
      const lng = longitudeExtractor(p);
      const intensity = intensityExtractor(p);
      return [lat, lng, intensity];
    });

    const heatLayer = (L as any).heatLayer(heatPoints, {
      radius,
      max,
      minOpacity,
      blur,
      gradient
    }).addTo(map);

    heatLayerRef.current = heatLayer;

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

  return null;
}

export default HeatmapLayer;