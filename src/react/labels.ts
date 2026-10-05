/** Every string the Themes page and the CircleXO return show. `{app}` is the host app's name. */
export interface ThemeLabels {
  title: string
  subtitle: string
  current: string
  fromProject: string
  colours: string
  reset: string
  resetConfirm: string
  resetToPlatform: string
  resetToOwn: string
  defaultName: string
  defaultBody: string
  inUse: string
  use: string
  premium: string
  customize: string
  demo: string
  customizeNote: string
  catalogTitle: string
  back: string
  studioTitle: string
  notConfigured: string
  pitch: string
  connect: string
  connected: string
  disconnect: string
  disconnectConfirm: string
  disconnectNote: string
  project: string
  noProjects: string
  useProject: string
  pull: string
  auto: string
  lastPull: string
  grantLost: string
  pullFailed: string
  connecting: string
  finishing: string
  denied: string
  error: string
}

const en: ThemeLabels = {
  title: "Themes",
  subtitle: "The look of your pages. Choose a Nasaq theme, or use the one you customised in Nasaq Studio.",
  current: "In use",
  fromProject: "From your Nasaq project, kept up to date",
  colours: "Theme colours, light and dark",
  reset: "Back to default",
  resetConfirm: "Go back to the default theme?",
  resetToPlatform: "Your pages take {app}'s default theme again.",
  resetToOwn: "Your pages go back to {app}'s own look and your brand colour.",
  defaultName: "{app} default",
  defaultBody: "{app}'s own look, in your brand colour.",
  inUse: "In use",
  use: "Use",
  premium: "Premium",
  customize: "Customise in Nasaq",
  demo: "Preview",
  customizeNote:
    "Customise a theme in Nasaq Studio, publish it, then connect CircleXO above and choose that project: {app} picks up every change you publish.",
  catalogTitle: "Nasaq themes",
  back: "Back to Themes",
  studioTitle: "Nasaq Studio",
  notConfigured: "CircleXO is not set up on {app} yet. You can still use any theme from the catalog below.",
  pitch: "Connect your CircleXO account to use the theme you customised in Nasaq Studio. {app} updates your pages whenever you publish.",
  connect: "Connect CircleXO",
  connected: "Connected",
  disconnect: "Disconnect",
  disconnectConfirm: "Disconnect CircleXO?",
  disconnectNote: "Your current theme stays; {app} stops following your Nasaq project.",
  project: "Nasaq project",
  noProjects: "Your organisation has no Nasaq projects yet. Create one in Nasaq Studio.",
  useProject: "Use this project",
  pull: "Update now",
  auto: "{app} follows this project: each theme you publish in Nasaq reaches your pages within minutes.",
  lastPull: "Last checked:",
  grantLost: "The CircleXO connection has expired. Disconnect, then connect again.",
  pullFailed: "The last update from Nasaq failed.",
  connecting: "Connecting CircleXO",
  finishing: "Finishing the connection…",
  denied: "CircleXO did not finish the sign-in. Go back and try again.",
  error: "Something went wrong. Try again.",
}

const ar: ThemeLabels = {
  title: "القوالب",
  subtitle: "مظهر صفحاتك. اختر قالبًا من نسق، أو استخدم القالب الذي خصّصته في نسق ستوديو.",
  current: "المستخدم حاليًا",
  fromProject: "من مشروعك على نسق، ويُحدَّث تلقائيًا",
  colours: "ألوان القالب، الفاتح والداكن",
  reset: "العودة إلى الافتراضي",
  resetConfirm: "العودة إلى القالب الافتراضي؟",
  resetToPlatform: "ستعود صفحاتك إلى قالب {app} الافتراضي.",
  resetToOwn: "ستعود صفحاتك إلى مظهر {app} ولون علامتك التجارية.",
  defaultName: "{app} الافتراضي",
  defaultBody: "مظهر {app} نفسه، بلون علامتك التجارية.",
  inUse: "مستخدم",
  use: "استخدام",
  premium: "مميز",
  customize: "تخصيص في نسق",
  demo: "معاينة",
  customizeNote: "خصّص القالب في نسق ستوديو وانشره، ثم اربط CircleXO أعلاه واختر ذلك المشروع: سيأخذ {app} كل تغيير تنشره.",
  catalogTitle: "قوالب نسق",
  back: "العودة إلى القوالب",
  studioTitle: "نسق ستوديو",
  notConfigured: "لم يتم إعداد CircleXO على {app} بعد. يمكنك استخدام أي قالب من القائمة أدناه.",
  pitch: "اربط حسابك على CircleXO لاستخدام القالب الذي خصّصته في نسق ستوديو. يحدّث {app} صفحاتك كلما نشرت.",
  connect: "ربط CircleXO",
  connected: "مرتبط",
  disconnect: "إلغاء الربط",
  disconnectConfirm: "إلغاء ربط CircleXO؟",
  disconnectNote: "سيبقى قالبك الحالي، لكن {app} سيتوقف عن متابعة مشروعك على نسق.",
  project: "مشروع نسق",
  noProjects: "لا توجد مشاريع نسق في مؤسستك بعد. أنشئ مشروعًا في نسق ستوديو.",
  useProject: "استخدام هذا المشروع",
  pull: "تحديث الآن",
  auto: "يتابع {app} هذا المشروع: كل قالب تنشره في نسق يصل إلى صفحاتك خلال دقائق.",
  lastPull: "آخر فحص:",
  grantLost: "انتهت صلاحية الربط مع CircleXO. ألغِ الربط ثم اربط مرة أخرى.",
  pullFailed: "فشل آخر تحديث من نسق.",
  connecting: "ربط CircleXO",
  finishing: "جارٍ إكمال الربط…",
  denied: "لم يُكمل CircleXO تسجيل الدخول. عد وحاول مرة أخرى.",
  error: "حدث خطأ. حاول مرة أخرى.",
}

/** The built-in labels in locale (Arabic for "ar", English otherwise), with `{app}` filled in. */
export function themeLabels(locale: string, app: string, override?: Partial<ThemeLabels>): ThemeLabels {
  const base = locale === "ar" ? ar : en
  const out = { ...base, ...override } as ThemeLabels
  for (const k of Object.keys(out) as (keyof ThemeLabels)[]) out[k] = out[k].replaceAll("{app}", app)
  return out
}
