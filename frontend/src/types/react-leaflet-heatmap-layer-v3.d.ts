declare module 'react-leaflet-heatmap-layer-v3' {
  import { Component, FC } from 'react';
  
  export interface HeatmapLayerProps {
    points: any[];
    longitudeExtractor: (point: any) => number;
    latitudeExtractor: (point: any) => number;
    intensityExtractor?: (point: any) => number;
    radius?: number;
    max?: number;
    minOpacity?: number;
    blur?: number;
    gradient?: {[key: string]: string};
  }

  // Original class component
  export class HeatmapLayer extends Component<HeatmapLayerProps> {}
  
  // Added functional component type for React 19 compatibility
  export const HeatmapLayerFC: FC<HeatmapLayerProps>;

  // Default export fallback
  const DefaultHeatmapLayer: typeof HeatmapLayer | typeof HeatmapLayerFC;
  export default DefaultHeatmapLayer;
}