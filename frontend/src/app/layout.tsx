import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import NavbarClient from "@/components/ui/navbar-client";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Air Quality Dashboard",
  description: "Monitor air quality and anomalies",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`min-h-screen bg-background font-sans antialiased ${inter.className} flex flex-col justify-center items-center`}>
        <NavbarClient />
        <main className="flex-1 container py-8">{children}</main>
      </body>
    </html>
  );
}
