'use client'

import { useState, useCallback } from 'react'
import Cropper from 'react-easy-crop'
import type { Area, Point } from 'react-easy-crop'
import { useTranslations } from 'next-intl'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Slider } from '@/components/ui/slider'
import { ZoomIn, ZoomOut, RotateCw } from 'lucide-react'

interface ImageCropperProps {
  image: string
  onCropComplete: (croppedImage: Blob) => void
  onCancel: () => void
  aspectRatio?: number
  open: boolean
}

// Create a cropped image from canvas
async function getCroppedImg(imageSrc: string, pixelCrop: Area, rotation = 0): Promise<Blob> {
  const image = await createImage(imageSrc)
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')

  if (!ctx) {
    throw new Error('Failed to get canvas context')
  }

  const maxSize = Math.max(image.width, image.height)
  const safeArea = 2 * ((maxSize / 2) * Math.sqrt(2))

  canvas.width = safeArea
  canvas.height = safeArea

  ctx.translate(safeArea / 2, safeArea / 2)
  ctx.rotate((rotation * Math.PI) / 180)
  ctx.translate(-safeArea / 2, -safeArea / 2)

  ctx.drawImage(image, safeArea / 2 - image.width * 0.5, safeArea / 2 - image.height * 0.5)

  const data = ctx.getImageData(0, 0, safeArea, safeArea)

  canvas.width = pixelCrop.width
  canvas.height = pixelCrop.height

  ctx.putImageData(
    data,
    Math.round(0 - safeArea / 2 + image.width * 0.5 - pixelCrop.x),
    Math.round(0 - safeArea / 2 + image.height * 0.5 - pixelCrop.y)
  )

  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (blob) {
          resolve(blob)
          /* c8 ignore next 3 */
        } else {
          reject(new Error('Failed to create blob'))
        }
      },
      'image/jpeg',
      0.95
    )
  })
}

function createImage(url: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.addEventListener('load', () => resolve(image))
    image.addEventListener('error', (error) => reject(error))
    image.setAttribute('crossOrigin', 'anonymous')
    image.src = url
  })
}

export function ImageCropper({
  image,
  onCropComplete,
  onCancel,
  aspectRatio = 1,
  open,
}: ImageCropperProps) {
  const t = useTranslations('imageCropper')
  const [crop, setCrop] = useState<Point>({ x: 0, y: 0 })
  const [zoom, setZoom] = useState(1)
  const [rotation, setRotation] = useState(0)
  const [croppedAreaPixels, setCroppedAreaPixels] = useState<Area | null>(null)
  const [isProcessing, setIsProcessing] = useState(false)

  const onCropChange = useCallback((newCrop: Point) => {
    setCrop(newCrop)
  }, [])

  const onZoomChange = useCallback((newZoom: number) => {
    setZoom(newZoom)
  }, [])

  const onCropCompleteHandler = useCallback((_croppedArea: Area, croppedAreaPixels: Area) => {
    setCroppedAreaPixels(croppedAreaPixels)
  }, [])

  const handleSave = useCallback(async () => {
    if (!croppedAreaPixels) return

    setIsProcessing(true)
    try {
      const croppedImage = await getCroppedImg(image, croppedAreaPixels, rotation)
      onCropComplete(croppedImage)
    } catch (error) {
      console.error('Error cropping image:', error)
    } finally {
      setIsProcessing(false)
    }
  }, [croppedAreaPixels, image, rotation, onCropComplete])

  const handleRotate = useCallback(() => {
    setRotation((prev) => (prev + 90) % 360)
  }, [])

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && onCancel()}>
      <DialogContent className="sm:max-w-[500px] p-0">
        <DialogHeader className="p-6 pb-0">
          <DialogTitle>{t('title')}</DialogTitle>
        </DialogHeader>

        <div className="relative h-[300px] bg-black">
          <Cropper
            image={image}
            crop={crop}
            zoom={zoom}
            rotation={rotation}
            aspect={aspectRatio}
            onCropChange={onCropChange}
            onZoomChange={onZoomChange}
            onCropComplete={onCropCompleteHandler}
            cropShape="round"
            showGrid={false}
          />
        </div>

        <div className="p-6 space-y-4">
          <div className="flex items-center gap-4">
            <ZoomOut className="h-4 w-4 text-muted-foreground" />
            <Slider
              value={[zoom]}
              min={1}
              max={3}
              step={0.1}
              onValueChange={([value]) => setZoom(value)}
              className="flex-1"
            />
            <ZoomIn className="h-4 w-4 text-muted-foreground" />
          </div>

          <div className="flex justify-center">
            <Button variant="outline" size="sm" onClick={handleRotate}>
              <RotateCw className="h-4 w-4 mr-2" />
              {t('rotate')}
            </Button>
          </div>
        </div>

        <DialogFooter className="p-6 pt-0">
          <Button variant="outline" onClick={onCancel} disabled={isProcessing}>
            {t('cancel')}
          </Button>
          <Button onClick={handleSave} disabled={isProcessing}>
            {isProcessing ? t('saving') : t('apply')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
