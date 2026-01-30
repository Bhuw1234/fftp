import { Canvas } from '@react-three/fiber'
import { Float, Text, RoundedBox, Line } from '@react-three/drei'
import * as THREE from 'three'

interface BlockProps {
    position: [number, number, number];
    color: string;
    label: string;
}

function FlowBlock({ position, color, label }: BlockProps) {
    return (
        <Float speed={2} rotationIntensity={0.1} floatIntensity={0.6} floatingRange={[-0.1, 0.1]}>
            <group position={position}>
                {/* Glass Box */}
                <RoundedBox args={[2, 1.2, 0.2]} radius={0.1} smoothness={4}>
                    <meshPhysicalMaterial
                        color={color}
                        transparent
                        opacity={0.1}
                        roughness={0}
                        metalness={0.1}
                        transmission={1}
                        thickness={2}
                    />
                </RoundedBox>

                {/* Glow Border */}
                <RoundedBox args={[2.05, 1.25, 0.15]} radius={0.1} smoothness={4}>
                    <meshBasicMaterial color={color} wireframe />
                </RoundedBox>

                <Text
                    position={[0, 0, 0.15]}
                    fontSize={0.2}
                    font="https://fonts.gstatic.com/s/inter/v12/UcCO3FwrK3iLTeHuS_fvQtMwCp50KnMw2boKoduKmMEVuLyfAZ9hjp-Ek-_EeA.woff"
                    color="white"
                    anchorX="center"
                    anchorY="middle"
                >
                    {label}
                </Text>
            </group>
        </Float>
    )
}

function Connection({ start, end }: { start: [number, number, number], end: [number, number, number] }) {
    // Simple line connection
    const points = [
        new THREE.Vector3(...start),
        new THREE.Vector3(...end)
    ]
    return (
        <Line
            points={points}
            color="white"
            opacity={0.2}
            transparent
            lineWidth={1}
        />
    )
}

export default function DataFlow() {
    return (
        <div className="w-full h-[400px] pointer-events-none">
            <Canvas camera={{ position: [0, 0, 6], fov: 45 }} gl={{ antialias: true }}>
                <ambientLight intensity={1} />
                <pointLight position={[10, 10, 10]} intensity={2} />

                <FlowBlock position={[-2.8, 0.8, 0]} color="#00ffff" label="Customer Data" />
                <FlowBlock position={[0, -0.8, 0]} color="#7c3aed" label="AI Analysis" />
                <FlowBlock position={[2.8, 0.8, 0]} color="#e5e7eb" label="Smart Ads" />

                {/* Static Lines as visual guides */}
                <Connection start={[-1.8, 0.2, 0]} end={[-0.8, -0.2, 0]} />
                <Connection start={[0.8, -0.2, 0]} end={[1.8, 0.2, 0]} />
            </Canvas>
        </div>
    )
}
