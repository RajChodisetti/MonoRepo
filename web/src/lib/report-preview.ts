export type PreviewCoordinates = {
  latitude?: number;
  longitude?: number;
};

export function parsePreviewCoordinates(
  rawLatitude: string | null,
  rawLongitude: string | null,
): PreviewCoordinates {
  const latitudeText = rawLatitude?.trim();
  const longitudeText = rawLongitude?.trim();
  if (!latitudeText || !longitudeText) return {};

  const latitude = Number(latitudeText);
  const longitude = Number(longitudeText);
  if (
    !Number.isFinite(latitude) ||
    !Number.isFinite(longitude) ||
    latitude < -90 ||
    latitude > 90 ||
    longitude < -180 ||
    longitude > 180
  ) {
    return {};
  }
  return { latitude, longitude };
}
