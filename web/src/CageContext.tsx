import { Waves } from 'lucide-react'

// Installation context only: this illustration is not a camera-position map.
export function CageContext({
  status,
  simulation,
}: {
  status: string
  simulation: boolean
}) {
  return (
    <section
      className="cage-context"
      aria-label="Sea-cage installation context"
    >
      <svg
        viewBox="0 0 260 125"
        aria-hidden="true"
        className="cage-illustration"
      >
        <defs>
          <linearGradient id="cage-water" x2="0" y2="1">
            <stop stopColor="#53cbb0" stopOpacity=".13" />
            <stop offset="1" stopColor="#286880" stopOpacity=".03" />
          </linearGradient>
          <pattern
            id="cage-net"
            width="12"
            height="12"
            patternUnits="userSpaceOnUse"
          >
            <path
              d="M0 0L12 12M12 0L0 12"
              stroke="#4b8e8a"
              strokeWidth=".5"
              opacity=".45"
            />
          </pattern>
          <clipPath id="net-shape">
            <path d="M38 39Q130 83 222 39L198 92Q130 131 62 92Z" />
          </clipPath>
        </defs>
        <path
          d="M5 48Q22 42 40 48T75 48M189 51Q205 44 224 50T257 50M11 73Q24 69 36 73M222 83Q238 78 250 83"
          fill="none"
          stroke="#477d83"
          opacity=".55"
        />
        <path
          d="M38 39Q130 83 222 39L198 92Q130 131 62 92Z"
          fill="url(#cage-water)"
          stroke="#437975"
          strokeWidth=".8"
        />
        <rect
          x="35"
          y="39"
          width="190"
          height="82"
          fill="url(#cage-net)"
          clipPath="url(#net-shape)"
        />
        <ellipse
          cx="130"
          cy="37"
          rx="94"
          ry="25"
          fill="#11282d"
          stroke="#79b6a5"
          strokeWidth="2"
        />
        <ellipse
          cx="130"
          cy="37"
          rx="86"
          ry="19"
          fill="url(#cage-water)"
          stroke="#477e7a"
          strokeWidth=".7"
        />
        <path
          d="M62 26L128 83L198 26"
          fill="none"
          stroke="#9fcaad"
          strokeWidth="1"
          strokeDasharray="3 3"
          opacity=".8"
        />
        <rect
          x="55"
          y="20"
          width="13"
          height="9"
          rx="2"
          fill="#183e34"
          stroke="#89dcc1"
        />
        <rect
          x="192"
          y="20"
          width="13"
          height="9"
          rx="2"
          fill="#213d50"
          stroke="#92bde0"
        />
        <rect
          x="123"
          y="78"
          width="13"
          height="9"
          rx="2"
          fill="#203a3e"
          stroke="#a0c6bd"
        />
        <path d="M136 80L140 78V87L136 85" fill="#80aca3" />
        <g fill="#659a9c" opacity=".65">
          <path d="M79 67q8-7 16 0q-8 7-16 0l-5 4v-8Z" />
          <path d="M163 81q8-6 16 0q-8 6-16 0l-5 4v-8Z" />
          <path d="M142 59q7-5 13 0q-6 5-13 0l-4 3v-6Z" />
        </g>
      </svg>
      <div className="cage-context-copy">
        <span className="eyebrow">AT THE CAGE EDGE</span>
        <h2>Above the water. In control below.</h2>
        <p>Sea-cage camera winch · independent X / Y drives</p>
        <span className="schematic-note">
          Installation schematic · not a camera-position map
        </span>
      </div>
      <div className="cage-status">
        <Waves size={20} />
        <span>{status}</span>
        <small>
          {simulation
            ? 'Simulated motor-axis feedback'
            : 'Position feedback from the drives'}
        </small>
      </div>
    </section>
  )
}
