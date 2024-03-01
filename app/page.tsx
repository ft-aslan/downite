import { useState } from 'react'
import { Button } from './components/ui/button'
import useSWR from 'swr'
import fetcher from './lib/fetcher'
function App() {
  const [count, setCount] = useState(0)
  const {data,error,isLoading} =useSWR('https://jsonplaceholder.typicode.com/todos/1',fetcher)
  return (
    <>
     <Button onClick={() => setCount((count) => count + 1)}>{count}</Button> 
    </>
  )
}

export default App
