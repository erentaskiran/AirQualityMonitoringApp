import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function debounce<F extends (...args: any[]) => any>(func: F, wait: number) {
  let timeout: ReturnType<typeof setTimeout> | null = null
  
  return function(...args: Parameters<F>) {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

export function MakeRequest(
  url: string,
  method: string,
  body: any = null,
  headers: HeadersInit = {}
): Promise<any> {
  const options: RequestInit = {
    method,
    headers: {
      "Content-Type": "application/json",
      ...headers,
    },
  }

  if (body) {
    options.body = JSON.stringify(body)
  }

  if (!url.startsWith("http")) {
    if (url.startsWith("http://localhost:8081/")) {
      url = url.replace("http://localhost:8081/", "")
    }
    
    if (url.startsWith("/")) {
      url = url.substring(1)
    }
    
    url = "http://localhost:8081/" + url
  }

  return fetch(url, options)
    .catch((error) => {
      console.error("Error:", error)
      throw error
    })
}