package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
)

//ColumnPrintToPDF print pdf
func ColumnPrintToPDF(aid int, filename string, cookies map[string]string) error {
	var buf []byte

	// disable chrome headless
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		//chromedp.Flag("headless", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		// context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Tasks{
			// 以 IPhone7 模拟设备浏览。
			chromedp.Emulate(device.IPhone7),
			enableLifeCycleEvents(),
			setCookies(cookies),
			navigateAndWaitFor(`https://time.geekbang.org/column/article/`+strconv.Itoa(aid), "networkIdle"),

			// 评论列表中评论可能折叠，循环打开折叠。
			chromedp.ActionFunc(func(ctx context.Context) error {
				s := `
					[...document.querySelectorAll('ul>li>div>div>div:nth-child(2)>span')].map(e=>e.click());
				`
				_, exp, err := runtime.Evaluate(s).Do(ctx)
				if err != nil {
					return err
				}

				if exp != nil {
					return exp
				}

				return nil
			}),
			chromedp.ActionFunc(func(ctx context.Context) error {
				// 1. 移除头部极客时间 Logo 标题
				// 2. 隐藏所有文字图标
				s := `
					document.querySelector(".shim") && (document.querySelector(".shim").style.display="none");
					document.querySelector(".main") && (document.querySelector(".main").style.display="none");
					[...document.querySelectorAll(".iconfont")].map(e=>e.style.display="none")
				`
				_, exp, err := runtime.Evaluate(s).Do(ctx)
				if err != nil {
					return err
				}

				if exp != nil {
					return exp
				}

				return nil
			}),
			// 调用 Chrome 内置 API 打印 PDF
			chromedp.ActionFunc(func(ctx context.Context) error {
				//time.Sleep(time.Millisecond * 1500)
				var err error
				buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
				return err
			}),
		},
	)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, buf, 0644)
}

func setCookies(cookies map[string]string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))

		for key, value := range cookies {
			success, err := network.SetCookie(key, value).WithExpires(&expr).WithDomain(".geekbang.org").WithHTTPOnly(true).Do(ctx)
			if err != nil {
				return err
			}

			if !success {
				return fmt.Errorf("could not set cookie %q to %q", key, value)
			}
		}
		return nil
	})
}

func enableLifeCycleEvents() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		err := page.Enable().Do(ctx)
		if err != nil {
			return err
		}

		return page.SetLifecycleEventsEnabled(true).Do(ctx)
	}
}

func navigateAndWaitFor(url string, eventName string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, _, _, err := page.Navigate(url).Do(ctx)
		if err != nil {
			return err
		}

		return waitFor(ctx, eventName)
	}
}

// waitFor blocks until eventName is received.
// Examples of events you can wait for:
//     init, DOMContentLoaded, firstPaint,
//     firstContentfulPaint, firstImagePaint,
//     firstMeaningfulPaintCandidate,
//     load, networkAlmostIdle, firstMeaningfulPaint, networkIdle
//
// This is not super reliable, I've already found incidental cases where
// networkIdle was sent before load. It's probably smart to see how
// puppeteer implements this exactly.
func waitFor(ctx context.Context, eventName string) error {
	ch := make(chan struct{})
	cctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(cctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *page.EventLifecycleEvent:
			if e.Name == eventName {
				cancel()
				close(ch)
			}
		}
	})

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
