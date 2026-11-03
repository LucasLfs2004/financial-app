export const featureFlags = {
  referenceBasis: process.env.NEXT_PUBLIC_ENABLE_REFERENCE_BASIS === "true",
  adjustmentMaintenance: process.env.NEXT_PUBLIC_ENABLE_INVOICE_ADJUSTMENT_MAINTENANCE === "true",
} as const;
