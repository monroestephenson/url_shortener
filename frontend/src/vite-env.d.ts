/// <reference types="vite/client" />

declare module '*.tsx' {
  const component: React.ComponentType
  export default component
} 