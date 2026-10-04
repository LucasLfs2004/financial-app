export const featureFlags = {
  referenceBasis: process.env.NEXT_PUBLIC_ENABLE_REFERENCE_BASIS === "true",
} as const;
