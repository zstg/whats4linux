import { useRef } from "react"
import { useGSAP } from "@gsap/react"
import gsap from "gsap"
import clsx from "clsx"

interface ToggleButtonProps {
  isEnabled: boolean
  onToggle: () => void
}

const ToggleButton = ({ isEnabled, onToggle }: ToggleButtonProps) => {
  const circleRef = useRef<HTMLDivElement>(null)

  useGSAP(() => {
    if (!circleRef.current) return

    gsap.to(circleRef.current, {
      x: isEnabled ? 20 : 0,
      duration: 0.6,
      ease: "elastic.out(1, 0.5)",
      overwrite: "auto",
    })
  }, [isEnabled])

  const handleClick = () => {
    onToggle()
  }

  return (
    <div
      className={clsx(
        "h-7 w-12 rounded-full flex items-center px-1 cursor-pointer shrink-0 transition-colors duration-300",
        isEnabled ? "bg-green-500" : "bg-gray-300 dark:bg-gray-600",
      )}
      onClick={handleClick}
    >
      <div className="size-5 rounded-full bg-white shadow-md translate-x-0" ref={circleRef} />
    </div>
  )
}

export default ToggleButton
