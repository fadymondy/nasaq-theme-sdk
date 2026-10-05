package theme

import "net/http"

// Error is a refusal worded in Arabic and English, with the HTTP status it maps to.
type Error struct {
	Status int
	EN, AR string
}

func (e *Error) Error() string { return e.EN }

// In is the message in locale ("ar" or anything else for English).
func (e *Error) In(locale string) string {
	if locale == "ar" {
		return e.AR
	}
	return e.EN
}

func refusal(status int, en, ar string) *Error { return &Error{Status: status, EN: en, AR: ar} }

// Refusals.
var (
	ErrUnknownTheme = refusal(http.StatusNotFound,
		"there is no such theme in Nasaq",
		"لا يوجد قالب بهذا الاسم في نسق")
	ErrNotForApp = refusal(http.StatusUnprocessableEntity,
		"this Nasaq theme does not cover this app's pages",
		"هذا القالب من نسق لا يغطي صفحات هذا التطبيق")
	ErrUnusable = refusal(http.StatusUnprocessableEntity,
		"this theme's colours cannot be used: its main colour is missing or unreadable",
		"لا يمكن استخدام ألوان هذا القالب: اللون الأساسي غير موجود أو غير مقروء")
	ErrNotConfigured = refusal(http.StatusServiceUnavailable,
		"CircleXO is not set up on this platform yet: the platform operator must add it in the platform settings",
		"لم يتم إعداد CircleXO على هذه المنصة بعد: يجب على مشغّل المنصة إضافته في إعدادات المنصة")
	ErrNoVault = refusal(http.StatusServiceUnavailable,
		"CircleXO cannot be connected: no key to seal its tokens is configured on this server",
		"لا يمكن ربط CircleXO: لم يتم إعداد مفتاح لحفظ رموزه على هذا الخادم")
	ErrState = refusal(http.StatusUnprocessableEntity,
		"this CircleXO sign-in has expired or belongs to someone else; start again",
		"انتهت صلاحية تسجيل الدخول إلى CircleXO أو أنه يخص شخصًا آخر، ابدأ من جديد")
	ErrRedirect = refusal(http.StatusUnprocessableEntity,
		"CircleXO can only return to this app",
		"يمكن لـ CircleXO العودة إلى هذا التطبيق فقط")
	ErrNoOrg = refusal(http.StatusUnprocessableEntity,
		"choose an organisation when you sign in to CircleXO",
		"اختر مؤسسة عند تسجيل الدخول إلى CircleXO")
	ErrNotLinked = refusal(http.StatusConflict,
		"this site is not connected to CircleXO",
		"هذا الموقع غير مرتبط بـ CircleXO")
	ErrNoProject = refusal(http.StatusConflict,
		"choose the Nasaq project this site uses first",
		"اختر أولًا مشروع نسق الذي يستخدمه هذا الموقع")
	ErrUnknownProject = refusal(http.StatusNotFound,
		"that project is not in your Nasaq organisation",
		"هذا المشروع غير موجود في مؤسستك على نسق")
	ErrNotInstalled = refusal(http.StatusConflict,
		"your CircleXO organisation has not installed Nasaq yet: add it from the CircleXO store, then try again",
		"لم تقم مؤسستك على CircleXO بتثبيت نسق بعد: أضفه من متجر CircleXO ثم حاول مرة أخرى")
	ErrGrantLost = refusal(http.StatusConflict,
		"the CircleXO connection has expired; connect again",
		"انتهت صلاحية الربط مع CircleXO، اربط الحساب مرة أخرى")
	ErrUpstream = refusal(http.StatusBadGateway,
		"Nasaq or CircleXO did not answer; try again in a moment",
		"لم يستجب نسق أو CircleXO، حاول مرة أخرى بعد قليل")
)

// All is every refusal (a bilingual test walks it).
var All = []*Error{ErrUnknownTheme, ErrNotForApp, ErrUnusable, ErrNotConfigured, ErrNoVault, ErrState,
	ErrRedirect, ErrNoOrg, ErrNotLinked, ErrNoProject, ErrUnknownProject, ErrNotInstalled, ErrGrantLost, ErrUpstream}
