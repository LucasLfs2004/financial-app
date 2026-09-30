import type { Metadata, Viewport } from "next";
import "./globals.css";
import { PwaRegister } from "./pwa-register";
import { ApiNotifications } from "@/components/api-notifications";

export const metadata: Metadata = {
  title: {
    default: "Projeção",
    template: "%s · Projeção",
  },
  description: "Planejamento financeiro simples, visual e sob seu controle.",
  applicationName: "Projeção",
  icons: {
    icon: [
      { url: "/icons/icon-192.png", sizes: "192x192", type: "image/png" },
      { url: "/icons/icon-512.png", sizes: "512x512", type: "image/png" },
    ],
    apple: [{ url: "/icons/apple-touch-icon.png", sizes: "180x180", type: "image/png" }],
  },
};

export const viewport: Viewport = {
  themeColor: "#07111f",
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="pt-BR" className="dark" data-scroll-behavior="smooth">
      <body>
        <PwaRegister />
        <ApiNotifications />
        {children}
      </body>
    </html>
  );
}
