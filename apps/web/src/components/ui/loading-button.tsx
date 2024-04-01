import { Loader2 } from "lucide-react"
import React from "react"
import { Button, ButtonProps } from "../ui/button"

type LoadingButtonProps = ButtonProps & {
  isLoading?: boolean
  children: React.ReactNode
}

export const LoadingButton = React.forwardRef<
  HTMLButtonElement,
  LoadingButtonProps
>(({ isLoading, children }, ref) => {
  if (isLoading) {
    return (
      <Button ref={ref} disabled={isLoading}>
        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
        {children}
      </Button>
    )
  }

  return (
    <Button ref={ref} disabled={isLoading}>
      {children}
    </Button>
  )
})
LoadingButton.displayName = "LoadingButton"
