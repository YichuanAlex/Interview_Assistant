declare const _default: __VLS_WithTemplateSlots<
  import('vue').DefineComponent<
    {
      position: {
        type: import('vue').PropType<import('../..').PanelPositionType>
        default: string
      }
      showZoom: {
        type: import('vue').PropType<boolean>
        default: boolean
      }
      showFitView: {
        type: import('vue').PropType<boolean>
        default: boolean
      }
      showInteractive: {
        type: import('vue').PropType<boolean>
        default: boolean
      }
      fitViewParams: {
        type: import('vue').PropType<import('../..').FitViewParams>
      }
    },
    {},
    unknown,
    {},
    {},
    import('vue').ComponentOptionsMixin,
    import('vue').ComponentOptionsMixin,
    {
      zoomIn: () => void
      zoomOut: () => void
      fitView: () => void
      interactionChange: (isInteractive: boolean) => void
    },
    string,
    import('vue').VNodeProps & import('vue').AllowedComponentProps & import('vue').ComponentCustomProps,
    Readonly<
      import('vue').ExtractPropTypes<{
        position: {
          type: import('vue').PropType<import('../..').PanelPositionType>
          default: string
        }
        showZoom: {
          type: import('vue').PropType<boolean>
          default: boolean
        }
        showFitView: {
          type: import('vue').PropType<boolean>
          default: boolean
        }
        showInteractive: {
          type: import('vue').PropType<boolean>
          default: boolean
        }
        fitViewParams: {
          type: import('vue').PropType<import('../..').FitViewParams>
        }
      }>
    > & {
      onZoomIn?: (() => any) | undefined
      onZoomOut?: (() => any) | undefined
      onFitView?: (() => any) | undefined
      onInteractionChange?: ((isInteractive: boolean) => any) | undefined
    },
    {
      position: import('../..').PanelPositionType
      showZoom: boolean
      showFitView: boolean
      showInteractive: boolean
    },
    {}
  >,
  {
    top?(_: {}): any
    'control-zoom-in'?(_: {}): any
    'icon-zoom-in'?(_: {}): any
    'control-zoom-out'?(_: {}): any
    'icon-zoom-out'?(_: {}): any
    'control-fit-view'?(_: {}): any
    'icon-fit-view'?(_: {}): any
    'control-interactive'?(_: {}): any
    'icon-unlock'?(_: {}): any
    'icon-lock'?(_: {}): any
    default?(_: {}): any
  }
>
export default _default
type __VLS_WithTemplateSlots<T, S> = T & {
  new (): {
    $slots: S
  }
}
