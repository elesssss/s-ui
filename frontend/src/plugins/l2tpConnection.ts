import type { Client, Link } from '@/types/clients'
import type { Inbound, L2TP } from '@/types/inbounds'

// Connection instructions belong to the user's QR dialog, rather than a
// proxy subscription. Read saved credentials without generating new ones.
export function l2tpConnectionLinks(client: Client, inbounds: Inbound[], t: (key: string) => string): Link[] {
  const credentials = client.config?.l2tp
  if (typeof credentials?.name !== 'string' || !credentials.name ||
      typeof credentials.password !== 'string' || !credentials.password) return []

  const selected = new Set(client.inbounds ?? [])
  return inbounds.flatMap(inbound => {
    if (inbound.type !== 'l2tp' || !selected.has(inbound.id)) return []
    const options = inbound as L2TP
    if (!options.public_ip || !options.psk) return []

    return [{
      type: 'local' as const,
      remark: `${inbound.tag} (L2TP/IPsec)`,
      uri: [
        'L2TP/IPsec (IKEv1 / PSK)',
        `${t('l2tp.publicIp')}: ${options.public_ip}`,
        `${t('l2tp.l2tpPort')}: 1701`,
        `${t('l2tp.ikePort')}: ${options.listen_port || 500}`,
        `${t('l2tp.nattPort')}: 4500`,
        `${t('l2tp.psk')}: ${options.psk}`,
        `${t('l2tp.username')}: ${credentials.name}`,
        `${t('l2tp.password')}: ${credentials.password}`,
      ].join('\n'),
    }]
  })
}
