"use client";

import Link from "next/link";
import { useState } from "react";
import { Menu } from "lucide-react";
import { Button } from "./button";
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  navigationMenuTriggerStyle,
} from "./navigation-menu";
import { cn } from "@/lib/utils";

export default function NavbarClient() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const toggleMobileMenu = () => {
    setMobileMenuOpen(!mobileMenuOpen);
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center justify-between">
        <Link href="/" className="flex items-center space-x-2">
          <span className="text-xl font-bold px-16">Air Quality Dashboard</span>
        </Link>
        
        {/* Desktop Navigation */}
        <NavigationMenu className="hidden md:flex flex-1 justify-center">
          <NavigationMenuList className="flex space-x-4">
            <NavigationMenuItem>
              <Link href="/" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Home
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/heatmap" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Heatmap
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/charts" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Charts
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/alerts" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Alerts
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
            <NavigationMenuItem>
              <Link href="/analysis" legacyBehavior passHref>
                <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                  Analysis
                </NavigationMenuLink>
              </Link>
            </NavigationMenuItem>
          </NavigationMenuList>
        </NavigationMenu>
        
        {/* Mobile Menu Button */}
        <div className="md:hidden">
          <Button 
            variant="ghost" 
            size="icon" 
            onClick={toggleMobileMenu} 
            aria-label="Toggle Navigation Menu"
          >
            <Menu className="h-6 w-6" />
          </Button>
        </div>
        
        {/* Empty div for spacing on desktop */}
        <div className="hidden md:block w-[180px]"></div>
      </div>
      
      {/* Mobile Navigation */}
      {mobileMenuOpen && (
        <div className="md:hidden border-t bg-background py-2">
          <nav className="container flex flex-col space-y-1">
            <Link href="/" className="px-4 py-2 hover:bg-accent rounded-md">
              Home
            </Link>
            <Link href="/heatmap" className="px-4 py-2 hover:bg-accent rounded-md">
              Heatmap
            </Link>
            <Link href="/charts" className="px-4 py-2 hover:bg-accent rounded-md">
              Charts
            </Link>
            <Link href="/alerts" className="px-4 py-2 hover:bg-accent rounded-md">
              Alerts
            </Link>
            <Link href="/analysis" className="px-4 py-2 hover:bg-accent rounded-md">
              Analysis
            </Link>
          </nav>
        </div>
      )}
    </header>
  );
}