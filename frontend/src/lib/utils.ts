import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
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

  // Fix the URL processing logic
  if (url.startsWith("http://localhost:8081/")) {
    url = url.replace("http://localhost:8081/", "")
  }
  
  if (url.startsWith("/")) {
    url = url.substring(1)
  }

  return fetch("http://localhost:8081/" + url, options)
    .then((response) => {
      return response
    })
    .catch((error) => {
      console.error("Error:", error)
      throw error
    })
}