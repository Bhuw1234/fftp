import { useRef } from 'react'
import { Canvas, useFrame } from '@react-three/fiber'
import { OrbitControls, Sphere, MeshDistortMaterial, Sparkles } from '@react-three/drei'
import * as THREE from 'three'

interface CoreProps {
    position?: [number, number, number]
}

function MovingCore({ position = [0, 0, 0] }: CoreProps) {
    const meshRef = useRef<THREE.Mesh>(null!)

    useFrame((state) => {
        if (meshRef.current) {
            meshRef.current.rotation.x = state.clock.getElapsedTime() * 0.2
            meshRef.current.rotation.y = state.clock.getElapsedTime() * 0.3
        }
    })

    // Icosahedron with distortion for "organic" tech feel
    return (
        <Sphere args={[1, 64, 64]} position={position} ref={meshRef}>
            <MeshDistortMaterial
                color="#00ffff"
                attach="material"
                distort={0.4}
                speed={1.5}
                roughness={0.2}
                metalness={0.9}
                emissive="#004444"
                emissiveIntensity={0.5}
            />
        </Sphere>
    )
}

export default function AICore() {
    return (
        <div className="absolute inset-0 z-0 h-full w-full pointer-events-auto">
            <Canvas camera={{ position: [0, 0, 4.5], fov: 45 }} gl={{ antialias: true, alpha: true }}>
                <ambientLight intensity={0.5} />
                <directionalLight position={[10, 10, 5]} intensity={1} color="#ffffff" />
                <pointLight position={[-10, -5, -5]} color="#7c3aed" intensity={5} distance={20} />

                <MovingCore />

                <Sparkles
                    count={300}
                    scale={8}
                    size={1.5}
                    speed={0.3}
                    opacity={0.6}
                    color="#00ffff"
                />

                {/* Adds floaty particles in background specifically */}
                <Sparkles
                    count={100}
                    scale={12}
                    size={3}
                    speed={0.2}
                    opacity={0.2}
                    color="#7c3aed"
                />

                <OrbitControls
                    enableZoom={false}
                    autoRotate
                    autoRotateSpeed={0.8}
                    enablePan={false}
                    enableDamping
                    dampingFactor={0.05}
                    maxPolarAngle={Math.PI / 1.5}
                    minPolarAngle={Math.PI / 2.5}
                />
            </Canvas>
        </div>
    )
}
