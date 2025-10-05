import { ModeToggle } from '@/components/ModeToggle'
import './App.css'
import { Outlet } from "react-router-dom"
import { Toaster } from '@/components/ui/sonner'

function App() {

  return (
    <>
      <div className="absolute top-4 right-4">
        <ModeToggle />
      </div>
      <main>
        <Outlet />
      </main>
      <Toaster />
    </>

  )
}

export default App
